package eventstore

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"heritage-tree-relocation-clearance/internal/domain"
)

const maxFrameBytes = 8 << 20

func (s *Store) Append(caseID string, expectedVersion uint64, idempotencyKey string, events []domain.Event) (AppendResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return AppendResult{}, errors.New("事件存储已关闭")
	}
	if strings.TrimSpace(caseID) == "" {
		return AppendResult{}, errors.New("caseID 不能为空")
	}
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 120 {
		return AppendResult{}, domain.Invalid("IDEMPOTENCY_KEY", "idempotencyKey 长度应为 8 至 120 个字符")
	}
	if len(events) == 0 {
		return AppendResult{}, errors.New("禁止追加空事件批次")
	}
	idemKey := caseID + "\x00" + idempotencyKey
	if result, ok := s.idempotency[idemKey]; ok {
		return AppendResult{Version: result.Version, Sequence: result.Sequence, Status: result.Status, Idempotent: true}, nil
	}
	currentVersion := uint64(0)
	if current := s.cases[caseID]; current != nil {
		currentVersion = current.Version
	}
	if currentVersion != expectedVersion {
		return AppendResult{}, &VersionConflictError{Expected: expectedVersion, Actual: currentVersion}
	}
	for _, event := range events {
		if event.CaseID != caseID {
			return AppendResult{}, fmt.Errorf("事件案卷编号 %s 与批次 %s 不一致", event.CaseID, caseID)
		}
	}
	candidate, err := cloneCase(s.cases[caseID])
	if err != nil {
		return AppendResult{}, err
	}
	if candidate == nil {
		candidate = domain.NewCaseState()
	}
	for _, event := range events {
		if err := candidate.Apply(event); err != nil {
			return AppendResult{}, fmt.Errorf("应用待提交事件: %w", err)
		}
	}
	f := frame{SchemaVersion: schemaVersion, Sequence: s.sequence + 1, PreviousDigest: s.lastDigest, CaseID: caseID, ExpectedVersion: expectedVersion, IdempotencyKey: idempotencyKey, Events: events}
	if current := s.cases[caseID]; current != nil && current.Credential == nil && candidate.Credential != nil {
		if candidate.Credential.EventSequence != f.Sequence {
			return AppendResult{}, errors.New("凭据事件边界与事件帧序号不一致")
		}
		if candidate.Credential.SerialNumber != s.nextSerial {
			return AppendResult{}, errors.New("凭据序号不是当前递增序号")
		}
		if _, exists := s.credentials[candidate.Credential.CredentialID]; exists {
			return AppendResult{}, errors.New("凭据编号已经存在")
		}
	}
	checksum, err := frameChecksum(f)
	if err != nil {
		return AppendResult{}, err
	}
	f.Checksum = checksum
	payload, err := json.Marshal(f)
	if err != nil {
		return AppendResult{}, err
	}
	if len(payload) > maxFrameBytes {
		return AppendResult{}, errors.New("事件帧超过 8 MiB 限制")
	}
	if err := appendPayload(s.logPath, payload); err != nil {
		return AppendResult{}, err
	}
	s.sequence, s.lastDigest = f.Sequence, frameDigest(payload)
	s.cases[caseID] = candidate
	result := idempotencyResult{CaseID: caseID, Version: candidate.Version, Sequence: f.Sequence, Status: candidate.Status}
	s.idempotency[idemKey] = result
	s.addRecords(f)
	if candidate.Credential != nil {
		s.credentials[candidate.Credential.CredentialID] = *candidate.Credential
		if candidate.Credential.SerialNumber >= s.nextSerial {
			s.nextSerial = candidate.Credential.SerialNumber + 1
		}
	}
	if err := s.writeSnapshotLocked(); err != nil {
		return AppendResult{}, fmt.Errorf("事件已提交但快照更新失败: %w", err)
	}
	return AppendResult{Version: candidate.Version, Sequence: f.Sequence, Status: candidate.Status}, nil
}

func appendPayload(path string, payload []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return fmt.Errorf("打开事件日志: %w", err)
	}
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(payload)))
	if _, err = file.Write(length); err == nil {
		_, err = file.Write(payload)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("写入事件帧: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("关闭事件日志: %w", closeErr)
	}
	return nil
}

func (s *Store) addRecords(f frame) {
	for index, event := range f.Events {
		record := EventRecord{Sequence: f.Sequence, EventIndex: index, CaseVersion: f.ExpectedVersion + uint64(index) + 1, Type: event.Type, CaseID: event.CaseID, OccurredAt: event.OccurredAt, Actor: event.Actor, Data: append([]byte(nil), event.Data...)}
		s.events[f.CaseID] = append(s.events[f.CaseID], record)
	}
}
