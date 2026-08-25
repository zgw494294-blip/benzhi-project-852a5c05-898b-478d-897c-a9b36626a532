package application

import (
	"encoding/json"

	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

type AuditItem struct {
	Sequence    uint64          `json:"sequence"`
	EventIndex  int             `json:"eventIndex"`
	CaseVersion uint64          `json:"caseVersion"`
	Type        string          `json:"type"`
	Actor       string          `json:"actor"`
	OccurredAt  string          `json:"occurredAt"`
	Data        json.RawMessage `json:"data"`
}

type AuditView struct {
	Case       *domain.RelocationCase      `json:"case"`
	Credential *domain.ClearanceCredential `json:"credential,omitempty"`
	Timeline   []AuditItem                 `json:"timeline"`
}

func (s *Service) GetCredential(credentialID string) (domain.ClearanceCredential, error) {
	return s.store.FindCredential(credentialID)
}

func (s *Service) GetAuditByCredential(credentialID string) (AuditView, error) {
	credential, err := s.store.FindCredential(credentialID)
	if err != nil {
		return AuditView{}, err
	}
	return s.GetAudit(credential.CaseID)
}

func (s *Service) GetAudit(caseID string) (AuditView, error) {
	var caseState *domain.RelocationCase
	var records []eventstore.EventRecord
	var queryErr error
	caseDone := make(chan struct{})
	auditDone := make(chan struct{})
	go func() {
		var err error
		caseState, err = s.store.LoadCase(caseID)
		if err != nil {
			queryErr = err
		}
		close(caseDone)
	}()
	go func() {
		var err error
		records, err = s.store.AuditTrail(caseID)
		if err != nil {
			queryErr = err
		}
		close(auditDone)
	}()
	<-caseDone
	<-auditDone
	if queryErr != nil {
		return AuditView{}, queryErr
	}
	items := make([]AuditItem, 0, len(records))
	for _, record := range records {
		items = append(items, auditItem(record))
	}
	return AuditView{Case: caseState, Credential: caseState.Credential, Timeline: items}, nil
}

func auditItem(record eventstore.EventRecord) AuditItem {
	return AuditItem{Sequence: record.Sequence, EventIndex: record.EventIndex, CaseVersion: record.CaseVersion, Type: record.Type, Actor: record.Actor, OccurredAt: record.OccurredAt.UTC().Format("2006-01-02T15:04:05.000000000Z"), Data: append([]byte(nil), record.Data...)}
}
