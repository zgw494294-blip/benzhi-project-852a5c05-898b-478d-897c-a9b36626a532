package multistorelogcoordination_test

import (
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

func TestSecondStoreCannotAppendFromStaleLogHead(t *testing.T) {
	directory := t.TempDir()
	firstStore, err := eventstore.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer firstStore.Close()
	secondStore, err := eventstore.Open(directory)
	if err != nil {
		return
	}
	defer secondStore.Close()

	start := make(chan struct{})
	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)
	go func() {
		<-start
		_, appendErr := firstStore.Append("CASE-STORE-A", 0, "create-store-a", createEvent(t, "CASE-STORE-A", "TREE-A01"))
		firstDone <- appendErr
	}()
	go func() {
		<-start
		if appendErr := <-firstDone; appendErr != nil {
			secondDone <- appendErr
			return
		}
		_, appendErr := secondStore.Append("CASE-STORE-B", 0, "create-store-b", createEvent(t, "CASE-STORE-B", "TREE-B01"))
		secondDone <- appendErr
	}()
	close(start)
	if err := <-secondDone; err == nil {
		t.Fatal("second store acknowledged an append from a stale log head")
	}
}

func createEvent(t *testing.T, caseID, treeCode string) []domain.Event {
	t.Helper()
	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	events, err := domain.CreateCase(domain.CreateCaseInput{
		CaseID: caseID, TreeCode: treeCode, ProtectionGrade: domain.GradeOne, Species: "香樟",
		TrunkDiameterCM: 80, CrownRadiusM: 6, PlannedWindowStart: now.Add(time.Hour),
		PlannedWindowEnd: now.Add(2 * time.Hour), Actor: "档案员甲", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return events
}
