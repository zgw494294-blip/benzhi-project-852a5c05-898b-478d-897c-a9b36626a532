package application

import (
	"crypto/sha256"
	"encoding/hex"

	"heritage-tree-relocation-clearance/internal/domain"
)

func (s *Service) CreateCase(command CreateCaseCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if command.CaseID == "" {
		command.CaseID = deterministicCaseID(command.IdempotencyKey)
	}
	context := WriteContext{ExpectedVersion: 0, IdempotencyKey: command.IdempotencyKey}
	if err := validateWriteContext(context); err != nil {
		return MutationResult{}, err
	}
	if s.store.CaseExists(command.CaseID) {
		// 让事件存储先识别完全相同的幂等重试；其他请求会得到版本冲突。
		return s.commit(&domain.RelocationCase{CaseID: command.CaseID}, context, []domain.Event{{CaseID: command.CaseID}})
	}
	events, err := domain.CreateCase(domain.CreateCaseInput{CaseID: command.CaseID, TreeCode: command.TreeCode, ProtectionGrade: command.ProtectionGrade, Species: command.Species, TrunkDiameterCM: command.TrunkDiameterCM, CrownRadiusM: command.CrownRadiusM, PlannedWindowStart: command.PlannedWindowStart, PlannedWindowEnd: command.PlannedWindowEnd, Actor: command.CreatedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	empty := domain.NewCaseState()
	empty.CaseID = command.CaseID
	return s.commit(empty, context, events)
}

// deterministicCaseID 在创建请求省略 caseID 时从幂等键派生稳定案卷编号，
// 使使用相同 idempotencyKey 的重试落到同一案卷，从而复用既有幂等机制返回首次结果。
func deterministicCaseID(idempotencyKey string) string {
	sum := sha256.Sum256([]byte(idempotencyKey))
	return "CASE-" + hex.EncodeToString(sum[:12])
}

func (s *Service) GetCase(caseID string) (*domain.RelocationCase, error) {
	return s.store.LoadCase(caseID)
}
