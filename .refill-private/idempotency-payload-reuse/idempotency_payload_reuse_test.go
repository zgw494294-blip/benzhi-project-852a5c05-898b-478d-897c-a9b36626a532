package idempotencypayloadreuse_test

import (
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

func TestReusedKeyWithDifferentPayloadIsRejected(t *testing.T) {
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	service := application.NewServiceWithClock(store, func() time.Time { return now })
	created, err := service.CreateCase(application.CreateCaseCommand{
		CaseID: "CASE-IDEM-PAYLOAD", TreeCode: "TREE-IDEM-02", ProtectionGrade: domain.GradeOne,
		Species: "银杏", TrunkDiameterCM: 80, CrownRadiusM: 6,
		PlannedWindowStart: now.Add(24 * time.Hour), PlannedWindowEnd: now.Add(48 * time.Hour),
		CreatedBy: "档案员甲", IdempotencyKey: "create-payload-case",
	})
	if err != nil {
		t.Fatal(err)
	}
	sharedKey := "survey-shared-key"
	first, err := service.RecordSurvey(application.RecordSurveyCommand{
		CaseID: "CASE-IDEM-PAYLOAD", WriteContext: application.WriteContext{ExpectedVersion: created.Version, IdempotencyKey: sharedKey},
		SurveyID: "SURVEY-N", Sector: "N", ProbeDepthCM: 90, CriticalRootCount: 3,
		ExposedRootRatio: 0.2, SoilMoisturePercent: 30, EvidenceRefs: []string{"photo://north"}, RecordedBy: "勘查员甲",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.RecordSurvey(application.RecordSurveyCommand{
		CaseID: "CASE-IDEM-PAYLOAD", WriteContext: application.WriteContext{ExpectedVersion: first.Version, IdempotencyKey: sharedKey},
		SurveyID: "SURVEY-E", Sector: "E", ProbeDepthCM: 100, CriticalRootCount: 4,
		ExposedRootRatio: 0.3, SoilMoisturePercent: 35, EvidenceRefs: []string{"photo://east"}, RecordedBy: "勘查员乙",
	})
	if !application.IsIdempotencyConflict(err) {
		state, loadErr := store.LoadCase("CASE-IDEM-PAYLOAD")
		if loadErr != nil {
			t.Fatal(loadErr)
		}
		t.Fatalf("different payload reused prior success: err=%v surveys=%v", err, state.Surveys)
	}
}
