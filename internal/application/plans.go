package application

import "heritage-tree-relocation-clearance/internal/domain"

func (s *Service) RevisePlan(command RevisePlanCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	fingerprint, err := fingerprintPayload("RevisePlan", command.CaseID, command.WriteContext.ExpectedVersion, planCommandPayload{
		PlanID: command.PlanID, CutBoundaryCM: command.CutBoundaryCM, RootBallRadiusCM: command.RootBallRadiusCM, RootBallDepthCM: command.RootBallDepthCM, SupportMethod: command.SupportMethod, MoistureMethod: command.MoistureMethod, MaxTransportMinutes: command.MaxTransportMinutes, MonitoringThresholds: command.MonitoringThresholds, PreparedBy: command.PreparedBy,
	})
	if err != nil {
		return MutationResult{}, err
	}
	if prior, ok, err := s.priorMutation(command.CaseID, command.WriteContext, fingerprint); err != nil {
		return MutationResult{}, err
	} else if ok {
		return prior, nil
	}
	caseState, err := s.loadForWrite(command.CaseID, command.WriteContext)
	if err != nil {
		return MutationResult{}, err
	}
	if command.PlanID == "" {
		command.PlanID = s.idGenerator("PLAN")
	}
	events, err := caseState.RevisePlan(domain.RevisePlanInput{PlanID: command.PlanID, CutBoundaryCM: command.CutBoundaryCM, RootBallRadiusCM: command.RootBallRadiusCM, RootBallDepthCM: command.RootBallDepthCM, SupportMethod: command.SupportMethod, MoistureMethod: command.MoistureMethod, MaxTransportMinutes: command.MaxTransportMinutes, MonitoringThresholds: command.MonitoringThresholds, PreparedBy: command.PreparedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, fingerprint, events)
}

type planCommandPayload struct {
	PlanID               string             `json:"planID,omitempty"`
	CutBoundaryCM        float64            `json:"cutBoundaryCM"`
	RootBallRadiusCM     float64            `json:"rootBallRadiusCM"`
	RootBallDepthCM      float64            `json:"rootBallDepthCM"`
	SupportMethod        string             `json:"supportMethod"`
	MoistureMethod       string             `json:"moistureMethod"`
	MaxTransportMinutes  int                `json:"maxTransportMinutes"`
	MonitoringThresholds map[string]float64 `json:"monitoringThresholds"`
	PreparedBy           string             `json:"preparedBy"`
}

func (s *Service) ReviewRisk(command ReviewRiskCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	fingerprint, err := fingerprintPayload("ReviewRisk", command.CaseID, command.WriteContext.ExpectedVersion, riskReviewCommandPayload{ReviewedBy: command.ReviewedBy})
	if err != nil {
		return MutationResult{}, err
	}
	if prior, ok, err := s.priorMutation(command.CaseID, command.WriteContext, fingerprint); err != nil {
		return MutationResult{}, err
	} else if ok {
		return prior, nil
	}
	caseState, err := s.loadForWrite(command.CaseID, command.WriteContext)
	if err != nil {
		return MutationResult{}, err
	}
	events, err := caseState.ReviewRisk(domain.ReviewRiskInput{Reviewer: command.ReviewedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, fingerprint, events)
}

type riskReviewCommandPayload struct {
	ReviewedBy string `json:"reviewedBy"`
}
