package restart_secondary_projections_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
	"heritage-tree-relocation-clearance/internal/httpapi"
)

func TestRestartRestoresCredentialAndAuditProjections(t *testing.T) {
	directory := t.TempDir()
	store, err := eventstore.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	credential := completeClearance(t, store)
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := eventstore.Open(directory)
	if err != nil {
		t.Fatalf("reopen event store: %v", err)
	}
	defer reopened.Close()
	api := httpapi.New(application.NewService(reopened), slog.New(slog.NewTextHandler(io.Discard, nil)))

	credentialResponse := httptest.NewRecorder()
	credentialRequest := httptest.NewRequest(http.MethodGet, "/api/v1/clearance-credentials/"+credential.CredentialID, nil)
	api.Handler().ServeHTTP(credentialResponse, credentialRequest)
	if credentialResponse.Code != http.StatusOK {
		t.Errorf("credential query after restart returned %d, body=%s", credentialResponse.Code, credentialResponse.Body.String())
	}

	auditResponse := httptest.NewRecorder()
	auditRequest := httptest.NewRequest(http.MethodGet, "/api/v1/relocation-cases/CASE-RESTART-1/audit", nil)
	api.Handler().ServeHTTP(auditResponse, auditRequest)
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("audit query after restart returned %d, body=%s", auditResponse.Code, auditResponse.Body.String())
	}
	var audit struct {
		Timeline []json.RawMessage `json:"timeline"`
	}
	if err := json.Unmarshal(auditResponse.Body.Bytes(), &audit); err != nil {
		t.Fatalf("decode audit response: %v", err)
	}
	if len(audit.Timeline) != 9 {
		t.Errorf("audit timeline after restart has %d events, want 9", len(audit.Timeline))
	}
	if next := reopened.PeekNextSerial(); next != 2 {
		t.Errorf("next credential serial after restart is %d, want 2", next)
	}
}

func completeClearance(t *testing.T, store *eventstore.Store) domain.ClearanceCredential {
	t.Helper()
	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	service := application.NewServiceWithClock(store, func() time.Time { return now })
	result, err := service.CreateCase(application.CreateCaseCommand{
		CaseID: "CASE-RESTART-1", TreeCode: "TREE-RESTART-1", ProtectionGrade: domain.GradeOne,
		Species: "银杏", TrunkDiameterCM: 80, CrownRadiusM: 5,
		PlannedWindowStart: now.Add(24 * time.Hour), PlannedWindowEnd: now.Add(48 * time.Hour),
		CreatedBy: "档案员甲", IdempotencyKey: "restart-create-001",
	})
	if err != nil {
		t.Fatal(err)
	}
	for index, sector := range []string{"N", "E", "S", "W"} {
		result, err = service.RecordSurvey(application.RecordSurveyCommand{
			CaseID: "CASE-RESTART-1", WriteContext: application.WriteContext{ExpectedVersion: result.Version, IdempotencyKey: "restart-survey-00" + string(rune('1'+index))},
			SurveyID: "SURVEY-" + sector, Sector: sector, ProbeDepthCM: 80, CriticalRootCount: 5,
			ExposedRootRatio: 0.2, SoilMoisturePercent: 30, EvidenceRefs: []string{"evidence-" + sector}, RecordedBy: "勘查员甲",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	result, err = service.RevisePlan(application.RevisePlanCommand{
		CaseID: "CASE-RESTART-1", WriteContext: application.WriteContext{ExpectedVersion: result.Version, IdempotencyKey: "restart-plan-001"},
		PlanID: "PLAN-RESTART-1", CutBoundaryCM: 550, RootBallRadiusCM: 500, RootBallDepthCM: 100,
		SupportMethod: "四向支撑", MoistureMethod: "持续保湿", MaxTransportMinutes: 120,
		MonitoringThresholds: map[string]float64{"rootVibrationMM": 5, "soilMoistureMinPercent": 30, "tiltDegrees": 3}, PreparedBy: "方案员甲",
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err = service.ReviewRisk(application.ReviewRiskCommand{CaseID: "CASE-RESTART-1", WriteContext: application.WriteContext{ExpectedVersion: result.Version, IdempotencyKey: "restart-risk-001"}, ReviewedBy: "复核员乙"})
	if err != nil {
		t.Fatal(err)
	}
	result, err = service.VerifySite(application.VerifySiteCommand{
		CaseID: "CASE-RESTART-1", WriteContext: application.WriteContext{ExpectedVersion: result.Version, IdempotencyKey: "restart-site-001"},
		WorkZoneReady: true, MachineryAccessReady: true, TemporaryCareReady: true, WeatherWindowSafe: true,
		Notes: "现场条件满足", VerifiedBy: "现场员丙",
	})
	if err != nil {
		t.Fatal(err)
	}
	credential, _, err := service.IssueCredential(application.IssueCredentialCommand{
		CaseID: "CASE-RESTART-1", WriteContext: application.WriteContext{ExpectedVersion: result.Version, IdempotencyKey: "restart-issue-001"}, IssuedBy: "签发员丁",
	})
	if err != nil {
		t.Fatal(err)
	}
	return credential
}
