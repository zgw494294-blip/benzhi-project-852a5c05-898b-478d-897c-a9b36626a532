package application

import "heritage-tree-relocation-clearance/internal/domain"

func (s *Service) RecordSurvey(command RecordSurveyCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if prior, ok, err := s.priorMutation(command.CaseID, command.WriteContext); err != nil {
		return MutationResult{}, err
	} else if ok {
		return prior, nil
	}
	caseState, err := s.loadForWrite(command.CaseID, command.WriteContext)
	if err != nil {
		return MutationResult{}, err
	}
	if command.SurveyID == "" {
		command.SurveyID = s.idGenerator("SURVEY")
	}
	events, err := caseState.RecordSurvey(domain.RecordSurveyInput{SurveyID: command.SurveyID, Sector: command.Sector, ProbeDepthCM: command.ProbeDepthCM, CriticalRootCount: command.CriticalRootCount, ExposedRootRatio: command.ExposedRootRatio, SoilMoisturePercent: command.SoilMoisturePercent, EvidenceRefs: command.EvidenceRefs, RecordedBy: command.RecordedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, events)
}
