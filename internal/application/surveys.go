package application

import "heritage-tree-relocation-clearance/internal/domain"

func (s *Service) RecordSurvey(command RecordSurveyCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	fingerprint, err := fingerprintPayload("RecordSurvey", command.CaseID, command.WriteContext.ExpectedVersion, surveyCommandPayload{
		SurveyID: command.SurveyID, Sector: command.Sector, ProbeDepthCM: command.ProbeDepthCM, CriticalRootCount: command.CriticalRootCount, ExposedRootRatio: command.ExposedRootRatio, SoilMoisturePercent: command.SoilMoisturePercent, EvidenceRefs: command.EvidenceRefs, RecordedBy: command.RecordedBy,
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
	if command.SurveyID == "" {
		command.SurveyID = s.idGenerator("SURVEY")
	}
	events, err := caseState.RecordSurvey(domain.RecordSurveyInput{SurveyID: command.SurveyID, Sector: command.Sector, ProbeDepthCM: command.ProbeDepthCM, CriticalRootCount: command.CriticalRootCount, ExposedRootRatio: command.ExposedRootRatio, SoilMoisturePercent: command.SoilMoisturePercent, EvidenceRefs: command.EvidenceRefs, RecordedBy: command.RecordedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, fingerprint, events)
}

type surveyCommandPayload struct {
	SurveyID            string   `json:"surveyID,omitempty"`
	Sector              string   `json:"sector"`
	ProbeDepthCM        float64  `json:"probeDepthCM"`
	CriticalRootCount   int      `json:"criticalRootCount"`
	ExposedRootRatio    float64  `json:"exposedRootRatio"`
	SoilMoisturePercent float64  `json:"soilMoisturePercent"`
	EvidenceRefs        []string `json:"evidenceRefs"`
	RecordedBy          string   `json:"recordedBy"`
}
