package application

import "heritage-tree-relocation-clearance/internal/domain"

func (s *Service) SubmitRemediation(command SubmitRemediationCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	fingerprint, err := fingerprintPayload("SubmitRemediation", command.CaseID, command.WriteContext.ExpectedVersion, submitRemediationCommandPayload{FindingID: command.FindingID, Evidence: command.Evidence, SubmittedBy: command.SubmittedBy})
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
	events, err := caseState.SubmitRemediation(domain.SubmitRemediationInput{FindingID: command.FindingID, Evidence: command.Evidence, SubmittedBy: command.SubmittedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, fingerprint, events)
}

type submitRemediationCommandPayload struct {
	FindingID   string   `json:"findingID"`
	Evidence    []string `json:"evidence"`
	SubmittedBy string   `json:"submittedBy"`
}

func (s *Service) ReviewRemediation(command ReviewRemediationCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	fingerprint, err := fingerprintPayload("ReviewRemediation", command.CaseID, command.WriteContext.ExpectedVersion, reviewRemediationCommandPayload{FindingID: command.FindingID, ReviewedBy: command.ReviewedBy, Decision: command.Decision})
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
	events, err := caseState.ReviewRemediation(domain.ReviewRemediationInput{FindingID: command.FindingID, ReviewedBy: command.ReviewedBy, Decision: command.Decision, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, fingerprint, events)
}

type reviewRemediationCommandPayload struct {
	FindingID   string                `json:"findingID"`
	ReviewedBy  string                `json:"reviewedBy"`
	Decision    domain.ReviewDecision `json:"decision"`
}
