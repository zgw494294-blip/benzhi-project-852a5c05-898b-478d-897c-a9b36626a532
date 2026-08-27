package auditdualerrorrace

import (
	"errors"
	"testing"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

type synchronizedFailureStore struct {
	arrived chan struct{}
	release chan struct{}
	failure error
}

func (s *synchronizedFailureStore) failTogether() error {
	s.arrived <- struct{}{}
	<-s.release
	return s.failure
}

func (s *synchronizedFailureStore) LoadCase(string) (*domain.RelocationCase, error) {
	return nil, s.failTogether()
}

func (s *synchronizedFailureStore) AuditTrail(string) ([]eventstore.EventRecord, error) {
	return nil, s.failTogether()
}

func (*synchronizedFailureStore) Append(string, uint64, string, []domain.Event) (eventstore.AppendResult, error) {
	panic("unexpected Append")
}

func (*synchronizedFailureStore) CaseExists(string) bool { panic("unexpected CaseExists") }

func (*synchronizedFailureStore) FindCredential(string) (domain.ClearanceCredential, error) {
	panic("unexpected FindCredential")
}

func (*synchronizedFailureStore) LastSequence() uint64 { panic("unexpected LastSequence") }

func (*synchronizedFailureStore) PeekNextSerial() uint64 { panic("unexpected PeekNextSerial") }

func (*synchronizedFailureStore) LookupIdempotency(string, string) (eventstore.AppendResult, bool) {
	panic("unexpected LookupIdempotency")
}

func TestConcurrentAuditFailuresDoNotRace(t *testing.T) {
	failure := errors.New("controlled audit storage failure")
	store := &synchronizedFailureStore{
		arrived: make(chan struct{}, 2),
		release: make(chan struct{}),
		failure: failure,
	}
	service := application.NewService(store)
	result := make(chan error, 1)

	go func() {
		_, err := service.GetAudit("CASE-MISSING")
		result <- err
	}()

	<-store.arrived
	<-store.arrived
	close(store.release)

	if err := <-result; !errors.Is(err, failure) {
		t.Fatalf("GetAudit error = %v, want controlled storage failure", err)
	}
}
