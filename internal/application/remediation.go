package application

import "heritage-tree-relocation-clearance/internal/domain"

func (s *Service) SubmitRemediation(command SubmitRemediationCommand) (MutationResult, error) {
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
	events, err := caseState.SubmitRemediation(domain.SubmitRemediationInput{FindingID: command.FindingID, Evidence: command.Evidence, SubmittedBy: command.SubmittedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, events)
}

func (s *Service) ReviewRemediation(command ReviewRemediationCommand) (MutationResult, error) {
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
	events, err := caseState.ReviewRemediation(domain.ReviewRemediationInput{FindingID: command.FindingID, ReviewedBy: command.ReviewedBy, Decision: command.Decision, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, events)
}
