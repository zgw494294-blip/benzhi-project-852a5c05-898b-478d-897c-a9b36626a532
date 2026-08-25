package eventstore

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"heritage-tree-relocation-clearance/internal/domain"
)

func (s *Store) recoverLog() error {
	file, err := os.Open(s.logPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("打开事件日志: %w", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var offset int64
	for {
		header := make([]byte, 4)
		n, readErr := io.ReadFull(reader, header)
		if readErr == io.EOF && n == 0 {
			break
		}
		if readErr != nil {
			return &CorruptionError{Offset: offset, Reason: "长度前缀被截断"}
		}
		length := binary.BigEndian.Uint32(header)
		if length == 0 || length > maxFrameBytes {
			return &CorruptionError{Offset: offset, Reason: "事件帧长度越界"}
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return &CorruptionError{Offset: offset + 4, Reason: "事件帧内容被截断"}
		}
		var f frame
		if err := json.Unmarshal(payload, &f); err != nil {
			return &CorruptionError{Offset: offset + 4, Reason: "事件帧 JSON 无效"}
		}
		if err := s.validateFrame(f, payload); err != nil {
			return &CorruptionError{Offset: offset, Reason: err.Error()}
		}
		if err := s.applyRecoveredFrame(f); err != nil {
			return &CorruptionError{Offset: offset, Reason: err.Error()}
		}
		s.sequence, s.lastDigest = f.Sequence, frameDigest(payload)
		offset += int64(4 + length)
	}
	return nil
}

func (s *Store) validateFrame(f frame, payload []byte) error {
	if f.SchemaVersion != schemaVersion {
		return fmt.Errorf("不支持 schemaVersion=%d", f.SchemaVersion)
	}
	if f.Sequence != s.sequence+1 {
		return fmt.Errorf("序号不连续，期望 %d，得到 %d", s.sequence+1, f.Sequence)
	}
	if f.PreviousDigest != s.lastDigest {
		return errors.New("前序摘要不匹配")
	}
	wanted, err := frameChecksum(f)
	if err != nil {
		return err
	}
	if wanted != f.Checksum {
		return errors.New("帧校验和不匹配")
	}
	if len(f.Events) == 0 {
		return errors.New("事件帧为空")
	}
	return nil
}

func (s *Store) applyRecoveredFrame(f frame) error {
	current := s.cases[f.CaseID]
	actual := uint64(0)
	if current != nil {
		actual = current.Version
	}
	if actual != f.ExpectedVersion {
		return fmt.Errorf("案卷版本链断裂，期望 %d，实际 %d", f.ExpectedVersion, actual)
	}
	candidate, err := cloneCase(current)
	if err != nil {
		return err
	}
	if candidate == nil {
		candidate = domain.NewCaseState()
	}
	for _, event := range f.Events {
		if event.CaseID != f.CaseID {
			return errors.New("事件案卷编号不匹配")
		}
		if err := candidate.Apply(event); err != nil {
			return err
		}
	}
	if current != nil && current.Credential == nil && candidate.Credential != nil {
		if candidate.Credential.EventSequence != f.Sequence {
			return errors.New("凭据事件边界与事件帧序号不一致")
		}
		if candidate.Credential.SerialNumber != s.nextSerial {
			return errors.New("凭据序号不连续")
		}
		if _, exists := s.credentials[candidate.Credential.CredentialID]; exists {
			return errors.New("凭据编号重复")
		}
	}
	s.cases[f.CaseID] = candidate
	s.idempotency[f.CaseID+"\x00"+f.IdempotencyKey] = idempotencyResult{CaseID: f.CaseID, Version: candidate.Version, Sequence: f.Sequence, Status: candidate.Status}
	return nil
}
