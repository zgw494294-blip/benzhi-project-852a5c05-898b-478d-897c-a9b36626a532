package eventstore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"heritage-tree-relocation-clearance/internal/domain"
)

func Open(dir string) (*Store, error) {
	if dir == "" {
		return nil, errors.New("事件存储目录不能为空")
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("创建事件存储目录: %w", err)
	}
	lockFile, err := lockDirectory(dir, ".store.lock")
	if err != nil {
		return nil, fmt.Errorf("事件存储目录已被占用或无法加锁: %w", err)
	}
	s := &Store{dir: dir, logPath: filepath.Join(dir, "events.log"), snapshotPath: filepath.Join(dir, "projection.snapshot.json"), nextSerial: 1, cases: map[string]*domain.RelocationCase{}, events: map[string][]EventRecord{}, credentials: map[string]domain.ClearanceCredential{}, idempotency: map[string]idempotencyResult{}}
	if err := s.recoverLog(); err != nil {
		_ = lockFile.Close()
		return nil, err
	}
	if err := s.verifySnapshot(); err != nil {
		_ = lockFile.Close()
		return nil, err
	}
	s.lockFile = lockFile
	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.lockFile != nil {
		_ = s.lockFile.Close()
		s.lockFile = nil
	}
	return nil
}

func (s *Store) Directory() string { return s.dir }

func cloneCase(source *domain.RelocationCase) (*domain.RelocationCase, error) {
	if source == nil {
		return nil, nil
	}
	b, err := jsonMarshal(source)
	if err != nil {
		return nil, err
	}
	var target domain.RelocationCase
	if err := jsonUnmarshal(b, &target); err != nil {
		return nil, err
	}
	return &target, nil
}

// 包装函数让序列化策略集中，未来升级 schema 时不会散落在查询代码中。
var jsonMarshal = func(v any) ([]byte, error) { return marshalJSON(v) }
var jsonUnmarshal = func(data []byte, v any) error { return unmarshalJSON(data, v) }
