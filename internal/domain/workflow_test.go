package domain

import (
	"testing"
	"time"
)

func TestCompleteWorkflowWithBlockingRemediation(t *testing.T) {
	now := time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC)
	events, err := CreateCase(CreateCaseInput{CaseID: "CASE-TEST-001", TreeCode: "TREE-001", ProtectionGrade: GradeOne, Species: "香樟", TrunkDiameterCM: 80, CrownRadiusM: 6, PlannedWindowStart: now.Add(24 * time.Hour), PlannedWindowEnd: now.Add(48 * time.Hour), Actor: "档案技术员", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	state := NewCaseState()
	applyAll(t, state, events)
	for index, sector := range RequiredSectors() {
		events, err = state.RecordSurvey(RecordSurveyInput{SurveyID: "SURVEY-" + sector, Sector: sector, ProbeDepthCM: 90, CriticalRootCount: 5, ExposedRootRatio: 0.2, SoilMoisturePercent: 30, EvidenceRefs: []string{"evidence-" + sector}, RecordedBy: "根系勘查员", Now: now.Add(time.Duration(index+1) * time.Minute)})
		if err != nil {
			t.Fatalf("sector %s: %v", sector, err)
		}
		applyAll(t, state, events)
	}
	if state.Status != StatusPlanning {
		t.Fatalf("status=%s", state.Status)
	}
	events, err = state.RevisePlan(RevisePlanInput{PlanID: "PLAN-001", CutBoundaryCM: 500, RootBallRadiusCM: 300, RootBallDepthCM: 100, SupportMethod: "四向柔性钢索支撑", MoistureMethod: "保湿布包覆并持续雾化", MaxTransportMinutes: 420, MonitoringThresholds: map[string]float64{"rootVibrationMM": 10, "soilMoistureMinPercent": 30, "tiltDegrees": 4}, PreparedBy: "方案编制员", Now: now.Add(10 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	applyAll(t, state, events)
	events, err = state.ReviewRisk(ReviewRiskInput{Reviewer: "风险负责人", Now: now.Add(11 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	applyAll(t, state, events)
	if state.Status != StatusRemediation || len(state.OpenBlockerIDs()) < 2 {
		t.Fatalf("findings=%v status=%s", state.Findings, state.Status)
	}
	for index, findingID := range state.OpenBlockerIDs() {
		events, err = state.SubmitRemediation(SubmitRemediationInput{FindingID: findingID, Evidence: []string{"整改照片与计算书"}, SubmittedBy: "整改经办员", Now: now.Add(time.Duration(20+index*2) * time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		applyAll(t, state, events)
		if _, err = state.ReviewRemediation(ReviewRemediationInput{FindingID: findingID, ReviewedBy: "整改经办员", Decision: DecisionAccept, Now: now.Add(time.Duration(21+index*2) * time.Minute)}); ErrorCode(err) != "INVALID_ROLE_SEPARATION" {
			t.Fatalf("expected role separation, got %v", err)
		}
		events, err = state.ReviewRemediation(ReviewRemediationInput{FindingID: findingID, ReviewedBy: "独立复核员", Decision: DecisionAccept, Now: now.Add(time.Duration(21+index*2) * time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		applyAll(t, state, events)
	}
	if state.Status != StatusSiteVerification {
		t.Fatalf("status=%s", state.Status)
	}
	events, err = state.VerifySite(VerifySiteInput{WorkZoneReady: true, MachineryAccessReady: true, TemporaryCareReady: true, WeatherWindowSafe: true, Notes: "条件满足", VerifiedBy: "现场核验员", Now: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	applyAll(t, state, events)
	if state.Status != StatusFrozen || state.Plans[1].FrozenAt == nil {
		t.Fatalf("freeze failed: %+v", state)
	}
	events, err = state.IssueCredential(IssueCredentialInput{CredentialID: "CLR-000000000001", SerialNumber: 1, EventSequence: 15, ContentDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", IssuedBy: "放行签发员", Now: now.Add(2 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	applyAll(t, state, events)
	if state.Status != StatusCleared || state.Credential == nil {
		t.Fatal("credential not applied")
	}
	if err := state.Apply(events[0]); ErrorCode(err) != "STATE_CONFLICT" {
		t.Fatalf("credential should be immutable: %v", err)
	}
}

func TestSurveyCoveragePreventsEarlyPlan(t *testing.T) {
	state := &RelocationCase{CaseID: "CASE-X", Status: StatusPreparing, TrunkDiameterCM: 50, Surveys: map[string]RootSurvey{}, Plans: map[int]ProtectionPlan{}, Findings: map[string]RiskFinding{}}
	_, err := state.RevisePlan(RevisePlanInput{})
	if ErrorCode(err) != "STATE_CONFLICT" {
		t.Fatalf("got %v", err)
	}
}

func applyAll(t *testing.T, state *RelocationCase, events []Event) {
	t.Helper()
	for _, event := range events {
		if err := state.Apply(event); err != nil {
			t.Fatal(err)
		}
	}
}
