package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const RiskRuleVersion = "risk-rules-2026.1"

type ReviewRiskInput struct {
	Reviewer string
	Now      time.Time
}

func (c *RelocationCase) ReviewRisk(in ReviewRiskInput) ([]Event, error) {
	if c.Status != StatusPlanning || c.CurrentPlanRevision == 0 {
		return nil, Conflict("当前状态不可提交风险审查")
	}
	if err := ValidateActor(in.Reviewer); err != nil {
		return nil, err
	}
	plan := c.Plans[c.CurrentPlanRevision]
	if plan.PreparedBy == in.Reviewer {
		return nil, Invalid("ROLE_SEPARATION", "风险审查人不得与方案编制人相同")
	}
	findings := c.evaluatePlan(plan)
	input := struct {
		RuleVersion string                `json:"ruleVersion"`
		Plan        ProtectionPlan        `json:"plan"`
		Surveys     map[string]RootSurvey `json:"surveys"`
	}{RiskRuleVersion, plan, c.Surveys}
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	e, err := NewEvent(EventRiskReviewCompleted, c.CaseID, in.Reviewer, in.Now, RiskReviewCompletedData{RuleVersion: RiskRuleVersion, InputDigest: hex.EncodeToString(sum[:]), Findings: findings})
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}

func (c *RelocationCase) evaluatePlan(plan ProtectionPlan) []RiskFinding {
	findings := make([]RiskFinding, 0)
	add := func(rule string, severity Severity, description string) {
		idSum := sha256.Sum256([]byte(c.CaseID + ":" + fmt.Sprint(plan.Revision) + ":" + rule))
		findings = append(findings, RiskFinding{FindingID: "RF-" + hex.EncodeToString(idSum[:6]), CaseID: c.CaseID, RuleCode: rule, Severity: severity, Description: description, Status: FindingOpen})
	}
	recommended := c.RecommendedRootBallRadius()
	if plan.RootBallRadiusCM < recommended {
		add("ROOT_BALL_BELOW_SURVEY", SeverityBlocker, fmt.Sprintf("土球半径 %.1fcm 小于勘查建议 %.1fcm", plan.RootBallRadiusCM, recommended))
	}
	if plan.CutBoundaryCM < plan.RootBallRadiusCM {
		add("CUT_BOUNDARY_INSIDE_BALL", SeverityBlocker, "断根边界不得位于土球半径以内")
	}
	if plan.MaxTransportMinutes > 360 {
		add("TRANSPORT_DURATION_HIGH", SeverityBlocker, "运输时限超过 360 分钟，需补充分段保湿与应急措施")
	}
	if plan.MonitoringThresholds["rootVibrationMM"] > 8 {
		add("VIBRATION_THRESHOLD_HIGH", SeverityBlocker, "根系振动阈值高于 8 毫米")
	}
	if plan.MonitoringThresholds["soilMoistureMinPercent"] < 25 {
		add("MOISTURE_THRESHOLD_LOW", SeverityWarning, "最低土壤含水率低于建议值 25%")
	}
	if plan.MonitoringThresholds["tiltDegrees"] > 5 {
		add("TILT_THRESHOLD_HIGH", SeverityWarning, "倾斜监测阈值高于建议值 5 度")
	}
	for _, survey := range c.Surveys {
		if survey.ExposedRootRatio > 0.5 {
			add("HIGH_EXPOSED_ROOT_"+survey.Sector, SeverityWarning, "方位 "+survey.Sector+" 的关键根暴露比例较高")
		}
		if survey.SoilMoisturePercent < 15 {
			add("DRY_SOIL_"+survey.Sector, SeverityBlocker, "方位 "+survey.Sector+" 土壤含水率过低")
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].RuleCode < findings[j].RuleCode })
	return findings
}
