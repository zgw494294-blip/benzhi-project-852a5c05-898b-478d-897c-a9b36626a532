package application

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

type Store interface {
	Append(caseID string, expectedVersion uint64, idempotencyKey string, events []domain.Event) (eventstore.AppendResult, error)
	AppendWithFingerprint(caseID string, expectedVersion uint64, idempotencyKey, fingerprint string, events []domain.Event) (eventstore.AppendResult, error)
	LoadCase(caseID string) (*domain.RelocationCase, error)
	CaseExists(caseID string) bool
	AuditTrail(caseID string) ([]eventstore.EventRecord, error)
	FindCredential(credentialID string) (domain.ClearanceCredential, error)
	LastSequence() uint64
	PeekNextSerial() uint64
	LookupIdempotency(caseID, idempotencyKey string) (eventstore.AppendResult, bool)
	LookupIdempotencyFingerprint(caseID, idempotencyKey string) (string, bool)
}

type Service struct {
	store       Store
	clock       func() time.Time
	idGenerator func(prefix string) string
	writeMu     sync.Mutex
}

func NewService(store Store) *Service {
	return &Service{store: store, clock: func() time.Time { return time.Now().UTC() }, idGenerator: randomID}
}

func NewServiceWithClock(store Store, clock func() time.Time) *Service {
	service := NewService(store)
	service.clock = clock
	return service
}

func randomID(prefix string) string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b)
}

func validateWriteContext(context WriteContext) error {
	if len(context.IdempotencyKey) < 8 || len(context.IdempotencyKey) > 120 {
		return domain.Invalid("IDEMPOTENCY_KEY", "idempotencyKey 长度应为 8 至 120 个字符")
	}
	return nil
}

func (s *Service) loadForWrite(caseID string, context WriteContext) (*domain.RelocationCase, error) {
	if err := validateWriteContext(context); err != nil {
		return nil, err
	}
	caseState, err := s.store.LoadCase(caseID)
	if err != nil {
		return nil, err
	}
	if caseState.Version != context.ExpectedVersion {
		return nil, &eventstore.VersionConflictError{Expected: context.ExpectedVersion, Actual: caseState.Version}
	}
	return caseState, nil
}

func fingerprintPayload(operation string, caseID string, expectedVersion uint64, payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("计算幂等指纹: %w", err)
	}
	sum := sha256.Sum256(append([]byte(operation+"\x00"+caseID+"\x00"+fmt.Sprintf("%d\x00", expectedVersion)), b...))
	return hex.EncodeToString(sum[:]), nil
}

func (s *Service) priorMutation(caseID string, context WriteContext, fingerprint string) (MutationResult, bool, error) {
	if err := validateWriteContext(context); err != nil {
		return MutationResult{}, false, err
	}
	storedFingerprint, ok := s.store.LookupIdempotencyFingerprint(caseID, context.IdempotencyKey)
	if !ok {
		return MutationResult{}, false, nil
	}
	if storedFingerprint != fingerprint {
		return MutationResult{}, false, &eventstore.IdempotencyPayloadConflictError{Key: context.IdempotencyKey}
	}
	result, ok := s.store.LookupIdempotency(caseID, context.IdempotencyKey)
	if !ok {
		return MutationResult{}, false, nil
	}
	return MutationResult{CaseID: caseID, Version: result.Version, Sequence: result.Sequence, Status: result.Status, Idempotent: true}, true, nil
}

func (s *Service) commit(caseState *domain.RelocationCase, context WriteContext, fingerprint string, events []domain.Event) (MutationResult, error) {
	result, err := s.store.AppendWithFingerprint(caseState.CaseID, context.ExpectedVersion, context.IdempotencyKey, fingerprint, events)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{CaseID: caseState.CaseID, Version: result.Version, Sequence: result.Sequence, Status: result.Status, Idempotent: result.Idempotent}, nil
}

func IsVersionConflict(err error) bool {
	var target *eventstore.VersionConflictError
	return errors.As(err, &target)
}
func IsIdempotencyConflict(err error) bool {
	var target *eventstore.IdempotencyConflictError
	if errors.As(err, &target) {
		return true
	}
	var payload *eventstore.IdempotencyPayloadConflictError
	return errors.As(err, &payload)
}
