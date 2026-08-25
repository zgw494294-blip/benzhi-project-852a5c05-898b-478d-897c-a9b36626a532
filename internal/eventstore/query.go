package eventstore

import (
	"sort"

	"heritage-tree-relocation-clearance/internal/domain"
)

func (s *Store) LoadCase(caseID string) (*domain.RelocationCase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	current := s.cases[caseID]
	if current == nil {
		return nil, domain.NotFound("案卷", caseID)
	}
	return cloneCase(current)
}

func (s *Store) CaseExists(caseID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cases[caseID] != nil
}

func (s *Store) AuditTrail(caseID string) ([]EventRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cases[caseID] == nil {
		return nil, domain.NotFound("案卷", caseID)
	}
	records := append([]EventRecord(nil), s.events[caseID]...)
	sort.Slice(records, func(i, j int) bool {
		if records[i].Sequence == records[j].Sequence {
			return records[i].EventIndex < records[j].EventIndex
		}
		return records[i].Sequence < records[j].Sequence
	})
	return records, nil
}

func (s *Store) FindCredential(credentialID string) (domain.ClearanceCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	credential, ok := s.credentials[credentialID]
	if !ok {
		return domain.ClearanceCredential{}, domain.NotFound("放行凭据", credentialID)
	}
	return credential, nil
}

func (s *Store) LastSequence() uint64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.sequence }

func (s *Store) PeekNextSerial() uint64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.nextSerial }

func (s *Store) LookupIdempotency(caseID, idempotencyKey string) (AppendResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.idempotency[caseID+"\x00"+idempotencyKey]
	if !ok {
		return AppendResult{}, false
	}
	return AppendResult{Version: result.Version, Sequence: result.Sequence, Status: result.Status, Idempotent: true}, true
}
