package generatedcaseididempotency_test

import (
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

func TestGeneratedCaseIDRetryReturnsOriginalMutation(t *testing.T) {
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	service := application.NewServiceWithClock(store, func() time.Time { return now })
	command := application.CreateCaseCommand{
		TreeCode:           "TREE-IDEM-01",
		ProtectionGrade:    domain.GradeOne,
		Species:            "香樟",
		TrunkDiameterCM:    80,
		CrownRadiusM:       6,
		PlannedWindowStart: now.Add(24 * time.Hour),
		PlannedWindowEnd:   now.Add(48 * time.Hour),
		CreatedBy:          "档案员甲",
		IdempotencyKey:     "generated-create-retry",
	}
	first, err := service.CreateCase(command)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.CreateCase(command)
	if err != nil {
		t.Fatal(err)
	}
	if second.CaseID != first.CaseID || !second.Idempotent || store.LastSequence() != 1 {
		t.Fatalf("retry created a second case: first=%+v second=%+v sequence=%d", first, second, store.LastSequence())
	}
}
