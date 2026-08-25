package application

import (
	"fmt"

	"heritage-tree-relocation-clearance/internal/domain"
)

func (s *Service) RevisePlan(command RevisePlanCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if prior, ok, err := s.priorMutation(command.CaseID, command.WriteContext); err != nil {
		return MutationResult{}, err
	} else if ok {
		return prior, nil
	}
	caseState, err := s.loadForWrite(command.CaseID, command.WriteContext)
	if err != nil {
		return MutationResult{}, fmt.Errorf("加载方案案卷失败: %v", err)
	}
	if command.PlanID == "" {
		command.PlanID = s.idGenerator("PLAN")
	}
	events, err := caseState.RevisePlan(domain.RevisePlanInput{PlanID: command.PlanID, CutBoundaryCM: command.CutBoundaryCM, RootBallRadiusCM: command.RootBallRadiusCM, RootBallDepthCM: command.RootBallDepthCM, SupportMethod: command.SupportMethod, MoistureMethod: command.MoistureMethod, MaxTransportMinutes: command.MaxTransportMinutes, MonitoringThresholds: command.MonitoringThresholds, PreparedBy: command.PreparedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, events)
}

func (s *Service) ReviewRisk(command ReviewRiskCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if prior, ok, err := s.priorMutation(command.CaseID, command.WriteContext); err != nil {
		return MutationResult{}, err
	} else if ok {
		return prior, nil
	}
	caseState, err := s.loadForWrite(command.CaseID, command.WriteContext)
	if err != nil {
		return MutationResult{}, fmt.Errorf("加载风险审查案卷失败: %v", err)
	}
	events, err := caseState.ReviewRisk(domain.ReviewRiskInput{Reviewer: command.ReviewedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, events)
}
