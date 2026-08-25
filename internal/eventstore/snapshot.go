package eventstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func (s *Store) writeSnapshotLocked() error {
	state := snapshot{SchemaVersion: schemaVersion, Sequence: s.sequence, LogDigest: s.lastDigest, NextSerial: s.nextSerial, Cases: s.cases, Idempotency: s.idempotency}
	checksum, err := snapshotChecksum(state)
	if err != nil {
		return err
	}
	state.Checksum = checksum
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(s.dir, ".projection-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	cleanup := func() { _ = temporary.Close(); _ = os.Remove(temporaryName) }
	if err := temporary.Chmod(0o640); err != nil {
		cleanup()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		cleanup()
		return err
	}
	if err := temporary.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryName)
		return err
	}
	if err := os.Rename(temporaryName, s.snapshotPath); err != nil {
		_ = os.Remove(temporaryName)
		return err
	}
	directory, err := os.Open(s.dir)
	if err != nil {
		return err
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

func (s *Store) verifySnapshot() error {
	payload, err := os.ReadFile(s.snapshotPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取投影快照: %w", err)
	}
	var state snapshot
	if err := json.Unmarshal(payload, &state); err != nil {
		return errors.New("投影快照 JSON 损坏")
	}
	if state.SchemaVersion != schemaVersion {
		return fmt.Errorf("投影快照 schemaVersion=%d 不受支持", state.SchemaVersion)
	}
	wanted, err := snapshotChecksum(state)
	if err != nil {
		return err
	}
	if wanted != state.Checksum {
		return errors.New("投影快照校验和不匹配")
	}
	if state.Sequence > s.sequence {
		return errors.New("投影快照超前于事件日志")
	}
	if state.Sequence == s.sequence && state.LogDigest != s.lastDigest {
		return errors.New("投影快照日志摘要不匹配")
	}
	if state.Sequence < s.sequence {
		return s.writeSnapshotLocked()
	}
	return nil
}
