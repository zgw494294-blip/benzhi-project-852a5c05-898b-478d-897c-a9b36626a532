package caseprojectionnestedalias_test

import (
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

func TestLoadedCaseMutationCannotAlterRiskReview(t *testing.T) {
	const caseID = "CASE-ALIAS-001"
	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	events := appendEvent(t, nil, domain.EventCaseCreated, caseID, "档案员甲", now, domain.CaseCreatedData{
		TreeCode: "TREE-ALIAS-001", ProtectionGrade: domain.GradeOne, Species: "银杏",
		TrunkDiameterCM: 80, CrownRadiusM: 6,
		PlannedWindowStart: now.Add(24 * time.Hour), PlannedWindowEnd: now.Add(48 * time.Hour),
	})
	for index, sector := range []string{"N", "E", "S", "W"} {
		survey := domain.RootSurvey{
			SurveyID: "SURVEY-" + sector, CaseID: caseID, Sector: sector,
			ProbeDepthCM: 100, CriticalRootCount: 2, ExposedRootRatio: 0.2,
			SoilMoisturePercent: 30, EvidenceRefs: []string{"photo-" + sector},
			RecommendedRootBallRadiusCM: 300, RecordedBy: "勘查员乙", RecordedAt: now.Add(time.Duration(index+1) * time.Minute),
		}
		events = appendEvent(t, events, domain.EventRootSurveyRecorded, caseID, "勘查员乙", survey.RecordedAt, domain.RootSurveyRecordedData{
			Survey: survey, CoverageComplete: index == 3, EvidenceCompleteness: float64(index+1) / 4,
		})
	}
	plan := domain.ProtectionPlan{
		PlanID: "PLAN-ALIAS-001", CaseID: caseID, Revision: 1,
		CutBoundaryCM: 360, RootBallRadiusCM: 340, RootBallDepthCM: 120,
		SupportMethod: "四向钢索支撑", MoistureMethod: "全程保湿覆盖", MaxTransportMinutes: 180,
		MonitoringThresholds: map[string]float64{"rootVibrationMM": 5, "soilMoistureMinPercent": 30, "tiltDegrees": 3},
		InputDigest:          "fixture", PreparedBy: "方案员丙", PreparedAt: now.Add(10 * time.Minute),
	}
	events = appendEvent(t, events, domain.EventProtectionPlanRevised, caseID, "方案员丙", plan.PreparedAt, domain.ProtectionPlanRevisedData{Plan: plan})
	if _, err := store.Append(caseID, 0, "alias-setup-001", events); err != nil {
		t.Fatal(err)
	}

	service := application.NewServiceWithClock(store, func() time.Time { return now.Add(20 * time.Minute) })
	loaded, err := service.GetCase(caseID)
	if err != nil {
		t.Fatal(err)
	}
	loaded.Plans[1].MonitoringThresholds["rootVibrationMM"] = 12

	_, err = service.ReviewRisk(application.ReviewRiskCommand{
		CaseID:       caseID,
		WriteContext: application.WriteContext{ExpectedVersion: 6, IdempotencyKey: "alias-review-001"},
		ReviewedBy:   "复核员丁",
	})
	if err != nil {
		t.Fatal(err)
	}
	after, err := service.GetCase(caseID)
	if err != nil {
		t.Fatal(err)
	}
	if finding, exists := findingByRule(after, "VIBRATION_THRESHOLD_HIGH"); exists {
		t.Fatalf("查询结果的嵌套 map 污染了持久化风险审查: %+v", finding)
	}
}

func appendEvent(t *testing.T, events []domain.Event, eventType, caseID, actor string, at time.Time, data any) []domain.Event {
	t.Helper()
	event, err := domain.NewEvent(eventType, caseID, actor, at, data)
	if err != nil {
		t.Fatal(err)
	}
	return append(events, event)
}

func findingByRule(caseState *domain.RelocationCase, rule string) (domain.RiskFinding, bool) {
	for _, finding := range caseState.Findings {
		if finding.RuleCode == rule {
			return finding, true
		}
	}
	return domain.RiskFinding{}, false
}
