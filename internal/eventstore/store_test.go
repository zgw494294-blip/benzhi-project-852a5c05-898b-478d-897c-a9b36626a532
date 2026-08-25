package eventstore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/domain"
)

func TestAppendIdempotencyAndRecovery(t *testing.T) {
	directory := t.TempDir()
	store, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	events := createEvents(t, "CASE-STORE-1")
	first, err := store.Append("CASE-STORE-1", 0, "create-key-001", events)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Append("CASE-STORE-1", 0, "create-key-001", []domain.Event{{CaseID: "invalid"}})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Idempotent || second.Version != first.Version || second.Status != domain.StatusPreparing {
		t.Fatalf("unexpected retry: %+v", second)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	state, err := reopened.LoadCase("CASE-STORE-1")
	if err != nil {
		t.Fatal(err)
	}
	if state.Version != 1 || state.TreeCode != "TREE-STORE-1" {
		t.Fatalf("bad recovered state: %+v", state)
	}
	if prior, ok := reopened.LookupIdempotency("CASE-STORE-1", "create-key-001"); !ok || prior.Sequence != 1 {
		t.Fatalf("missing idempotency: %+v %t", prior, ok)
	}
}

func TestRejectsTruncatedAndTamperedLog(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{name: "truncated", mutate: func(data []byte) []byte { return data[:len(data)-3] }},
		{name: "tampered", mutate: func(data []byte) []byte {
			copied := append([]byte(nil), data...)
			copied[len(copied)-8] ^= 1
			return copied
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			directory := t.TempDir()
			store, err := Open(directory)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = store.Append("CASE-CORRUPT", 0, "create-key-corrupt", createEvents(t, "CASE-CORRUPT")); err != nil {
				t.Fatal(err)
			}
			_ = store.Close()
			logPath := filepath.Join(directory, "events.log")
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(logPath, testCase.mutate(data), 0o640); err != nil {
				t.Fatal(err)
			}
			_, err = Open(directory)
			var corruption *CorruptionError
			if !errors.As(err, &corruption) {
				t.Fatalf("expected corruption, got %v", err)
			}
		})
	}
}

func TestRejectsTamperedSnapshot(t *testing.T) {
	directory := t.TempDir()
	store, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Append("CASE-SNAPSHOT", 0, "create-key-snapshot", createEvents(t, "CASE-SNAPSHOT")); err != nil {
		t.Fatal(err)
	}
	_ = store.Close()
	path := filepath.Join(directory, "projection.snapshot.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-5] ^= 1
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory); err == nil {
		t.Fatal("tampered snapshot accepted")
	}
}

func createEvents(t *testing.T, caseID string) []domain.Event {
	t.Helper()
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	events, err := domain.CreateCase(domain.CreateCaseInput{CaseID: caseID, TreeCode: "TREE-STORE-1", ProtectionGrade: domain.GradeOne, Species: "银杏", TrunkDiameterCM: 80, CrownRadiusM: 5, PlannedWindowStart: now.Add(time.Hour), PlannedWindowEnd: now.Add(2 * time.Hour), Actor: "档案员甲", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	return events
}
