package snapshoterrorcommit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

func TestSnapshotFailureCannotReturnErrorAfterCommit(t *testing.T) {
	directory := t.TempDir()
	store, err := eventstore.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := os.Mkdir(filepath.Join(directory, "projection.snapshot.json"), 0o750); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	events, err := domain.CreateCase(domain.CreateCaseInput{
		CaseID: "CASE-SNAPSHOT-FAIL", TreeCode: "TREE-SNAP", ProtectionGrade: domain.GradeOne,
		Species: "银杏", TrunkDiameterCM: 80, CrownRadiusM: 6,
		PlannedWindowStart: now.Add(time.Hour), PlannedWindowEnd: now.Add(2 * time.Hour), Actor: "档案员甲", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, appendErr := store.Append("CASE-SNAPSHOT-FAIL", 0, "snapshot-fail-key", events)
	committed := store.CaseExists("CASE-SNAPSHOT-FAIL") && store.LastSequence() == 1
	if appendErr != nil && committed {
		t.Fatalf("append returned error after durable mutation became visible: %v", appendErr)
	}
}
