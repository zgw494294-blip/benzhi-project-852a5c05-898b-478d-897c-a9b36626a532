package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"heritage-tree-relocation-clearance/internal/domain"
)

func (s *Service) VerifySite(command VerifySiteCommand) (MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if prior, ok, err := s.priorMutation(command.CaseID, command.WriteContext); err != nil {
		return MutationResult{}, err
	} else if ok {
		return prior, nil
	}
	caseState, err := s.loadForWrite(command.CaseID, command.WriteContext)
	if err != nil {
		return MutationResult{}, fmt.Errorf("加载现场核验案卷失败: %v", err)
	}
	events, err := caseState.VerifySite(domain.VerifySiteInput{WorkZoneReady: command.WorkZoneReady, MachineryAccessReady: command.MachineryAccessReady, TemporaryCareReady: command.TemporaryCareReady, WeatherWindowSafe: command.WeatherWindowSafe, Notes: command.Notes, VerifiedBy: command.VerifiedBy, Now: s.clock()})
	if err != nil {
		return MutationResult{}, err
	}
	return s.commit(caseState, command.WriteContext, events)
}

func (s *Service) IssueCredential(command IssueCredentialCommand) (domain.ClearanceCredential, MutationResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if prior, ok, err := s.priorMutation(command.CaseID, command.WriteContext); err != nil {
		return domain.ClearanceCredential{}, MutationResult{}, err
	} else if ok {
		caseState, loadErr := s.store.LoadCase(command.CaseID)
		if loadErr != nil {
			return domain.ClearanceCredential{}, MutationResult{}, loadErr
		}
		if caseState.Credential == nil {
			return domain.ClearanceCredential{}, MutationResult{}, domain.Conflict("幂等签发结果缺少凭据")
		}
		return *caseState.Credential, prior, nil
	}
	caseState, err := s.loadForWrite(command.CaseID, command.WriteContext)
	if err != nil {
		return domain.ClearanceCredential{}, MutationResult{}, fmt.Errorf("加载凭据签发案卷失败: %v", err)
	}
	serial := s.store.PeekNextSerial()
	sequence := s.store.LastSequence() + 1
	credentialID := fmt.Sprintf("CLR-%012d", serial)
	digestPayload := struct {
		CaseID              string                `json:"caseID"`
		TreeCode            string                `json:"treeCode"`
		Plan                domain.ProtectionPlan `json:"plan"`
		SiteChecklistDigest string                `json:"siteChecklistDigest"`
		ReviewInputDigest   string                `json:"reviewInputDigest"`
		Serial              uint64                `json:"serial"`
		EventSequence       uint64                `json:"eventSequence"`
	}{caseState.CaseID, caseState.TreeCode, caseState.Plans[caseState.FrozenPlanRevision], caseState.SiteChecklistDigest, caseState.ReviewInputDigest, serial, sequence}
	payload, err := json.Marshal(digestPayload)
	if err != nil {
		return domain.ClearanceCredential{}, MutationResult{}, err
	}
	sum := sha256.Sum256(payload)
	events, err := caseState.IssueCredential(domain.IssueCredentialInput{CredentialID: credentialID, SerialNumber: serial, EventSequence: sequence, ContentDigest: hex.EncodeToString(sum[:]), IssuedBy: command.IssuedBy, Now: s.clock()})
	if err != nil {
		return domain.ClearanceCredential{}, MutationResult{}, err
	}
	result, err := s.commit(caseState, command.WriteContext, events)
	if err != nil {
		return domain.ClearanceCredential{}, MutationResult{}, err
	}
	credential, err := s.store.FindCredential(credentialID)
	return credential, result, err
}
