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
	s := &Store{dir: dir, logPath: filepath.Join(dir, "events.log"), snapshotPath: filepath.Join(dir, "projection.snapshot.json"), nextSerial: 1, cases: map[string]*domain.RelocationCase{}, events: map[string][]EventRecord{}, credentials: map[string]domain.ClearanceCredential{}, idempotency: map[string]idempotencyResult{}}
	if err := s.recoverLog(); err != nil {
		return nil, err
	}
	if err := s.verifySnapshot(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func (s *Store) Directory() string { return s.dir }

func cloneCase(source *domain.RelocationCase) (*domain.RelocationCase, error) {
	if source == nil {
		return nil, nil
	}
	target := *source
	target.Surveys = make(map[string]domain.RootSurvey, len(source.Surveys))
	for sector, survey := range source.Surveys {
		target.Surveys[sector] = survey
	}
	target.Plans = make(map[int]domain.ProtectionPlan, len(source.Plans))
	for revision, plan := range source.Plans {
		target.Plans[revision] = plan
	}
	target.Findings = make(map[string]domain.RiskFinding, len(source.Findings))
	for findingID, finding := range source.Findings {
		target.Findings[findingID] = finding
	}
	return &target, nil
}

// 包装函数让序列化策略集中，未来升级 schema 时不会散落在查询代码中。
var jsonMarshal = func(v any) ([]byte, error) { return marshalJSON(v) }
var jsonUnmarshal = func(data []byte, v any) error { return unmarshalJSON(data, v) }
