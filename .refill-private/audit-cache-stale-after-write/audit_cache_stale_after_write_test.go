package audit_cache_stale_after_write_test

import (
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

func TestAuditCacheInvalidatedAfterCommittedWrite(t *testing.T) {
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC)
	service := application.NewServiceWithClock(store, func() time.Time {
		return now
	})
	created, err := service.CreateCase(application.CreateCaseCommand{
		CaseID:             "CASE-AUDIT-CACHE",
		TreeCode:           "TREE-AUDIT-CACHE",
		ProtectionGrade:    domain.GradeOne,
		Species:            "银杏",
		TrunkDiameterCM:    80,
		CrownRadiusM:       5,
		PlannedWindowStart: now.Add(24 * time.Hour),
		PlannedWindowEnd:   now.Add(48 * time.Hour),
		CreatedBy:          "档案员甲",
		IdempotencyKey:     "create-audit-cache-case",
	})
	if err != nil {
		t.Fatal(err)
	}

	first, err := service.GetAudit(created.CaseID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Case.Version != 1 || len(first.Timeline) != 1 {
		t.Fatalf("unexpected initial audit: version=%d timeline=%d", first.Case.Version, len(first.Timeline))
	}
	first.Case.TreeCode = "CALLER-MUTATED-TREE"
	first.Timeline[0].Type = "CallerMutatedEvent"
	cachedAgain, err := service.GetAudit(created.CaseID)
	if err != nil {
		t.Fatal(err)
	}
	aliasPoisoned := cachedAgain.Case.TreeCode == "CALLER-MUTATED-TREE" || cachedAgain.Timeline[0].Type == "CallerMutatedEvent"

	now = now.Add(time.Minute)
	mutation, err := service.RecordSurvey(application.RecordSurveyCommand{
		CaseID: created.CaseID,
		WriteContext: application.WriteContext{
			ExpectedVersion: created.Version,
			IdempotencyKey:  "record-north-survey-cache",
		},
		SurveyID:            "SURVEY-NORTH-CACHE",
		Sector:              "N",
		ProbeDepthCM:        90,
		CriticalRootCount:   4,
		ExposedRootRatio:    0.2,
		SoilMoisturePercent: 30,
		EvidenceRefs:        []string{"evidence://north/cache"},
		RecordedBy:          "勘查员乙",
	})
	if err != nil {
		t.Fatal(err)
	}
	if mutation.Version != 2 || mutation.Sequence != 2 {
		t.Fatalf("unexpected committed mutation: %+v", mutation)
	}

	second, err := service.GetAudit(created.CaseID)
	if err != nil {
		t.Fatal(err)
	}
	staleAfterWrite := second.Case.Version != mutation.Version || len(second.Timeline) != 2
	if aliasPoisoned || staleAfterWrite {
		t.Fatalf("TestAuditCacheInvalidatedAfterCommittedWrite: aliasPoisoned=%t committed version=%d sequence=%d, audit version=%d timeline=%d", aliasPoisoned, mutation.Version, mutation.Sequence, second.Case.Version, len(second.Timeline))
	}
}
