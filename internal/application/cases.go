package application

import (
	"heritage-tree-relocation-clearance/internal/domain"
)

func (s *Service) CreateCase(command CreateCaseCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if command.CaseID == "" {
		command.CaseID = s.idGenerator("CASE")
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

func (s *Service) GetCase(caseID string) (*domain.RelocationCase, error) {
	return s.store.LoadCase(caseID)
}
