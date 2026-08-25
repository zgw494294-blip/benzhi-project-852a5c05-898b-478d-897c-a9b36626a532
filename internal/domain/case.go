package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type CreateCaseInput struct {
	CaseID             string
	TreeCode           string
	ProtectionGrade    ProtectionGrade
	Species            string
	TrunkDiameterCM    float64
	CrownRadiusM       float64
	PlannedWindowStart time.Time
	PlannedWindowEnd   time.Time
	Actor              string
	Now                time.Time
}

func CreateCase(in CreateCaseInput) ([]Event, error) {
	if in.CaseID == "" {
		return nil, Invalid("CASE_ID", "案卷编号不能为空")
	}
	if err := ValidateActor(in.Actor); err != nil {
		return nil, err
	}
	if err := ValidateCaseProfile(in.TreeCode, in.ProtectionGrade, in.Species, in.TrunkDiameterCM, in.CrownRadiusM, in.PlannedWindowStart, in.PlannedWindowEnd); err != nil {
		return nil, err
	}
	if !in.PlannedWindowStart.After(in.Now) {
		return nil, Invalid("PLANNED_WINDOW", "计划移栽窗口必须晚于建档时间")
	}
	data := CaseCreatedData{TreeCode: in.TreeCode, ProtectionGrade: in.ProtectionGrade, Species: in.Species, TrunkDiameterCM: in.TrunkDiameterCM, CrownRadiusM: in.CrownRadiusM, PlannedWindowStart: in.PlannedWindowStart.UTC(), PlannedWindowEnd: in.PlannedWindowEnd.UTC()}
	e, err := NewEvent(EventCaseCreated, in.CaseID, in.Actor, in.Now, data)
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}

func (c *RelocationCase) Apply(event Event) error {
	if c.Surveys == nil {
		c.Surveys = map[string]RootSurvey{}
	}
	if c.Plans == nil {
		c.Plans = map[int]ProtectionPlan{}
	}
	if c.Findings == nil {
		c.Findings = map[string]RiskFinding{}
	}
	switch event.Type {
	case EventCaseCreated:
		var d CaseCreatedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		c.CaseID, c.TreeCode, c.ProtectionGrade, c.Species = event.CaseID, d.TreeCode, d.ProtectionGrade, d.Species
		c.TrunkDiameterCM, c.CrownRadiusM = d.TrunkDiameterCM, d.CrownRadiusM
		c.PlannedWindowStart, c.PlannedWindowEnd = d.PlannedWindowStart, d.PlannedWindowEnd
		c.Status, c.CreatedAt = StatusPreparing, event.OccurredAt
	case EventRootSurveyRecorded:
		var d RootSurveyRecordedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		c.Surveys[d.Survey.Sector] = d.Survey
		if d.CoverageComplete {
			c.Status = StatusPlanning
		}
	case EventProtectionPlanRevised:
		var d ProtectionPlanRevisedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		c.Plans[d.Plan.Revision] = d.Plan
		c.CurrentPlanRevision = d.Plan.Revision
		c.Status = StatusPlanning
	case EventRiskReviewCompleted:
		var d RiskReviewCompletedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		c.Findings = map[string]RiskFinding{}
		blockers := 0
		for _, finding := range d.Findings {
			c.Findings[finding.FindingID] = finding
			if finding.Severity == SeverityBlocker {
				blockers++
			}
		}
		c.ReviewRuleVersion, c.ReviewInputDigest = d.RuleVersion, d.InputDigest
		if blockers > 0 {
			c.Status = StatusRemediation
		} else {
			c.Status = StatusSiteVerification
		}
	case EventRemediationSubmitted:
		var d RemediationSubmittedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		f := c.Findings[d.FindingID]
		f.Status, f.RemediationEvidence, f.SubmittedBy = FindingSubmitted, append([]string(nil), d.Evidence...), d.SubmittedBy
		c.Findings[d.FindingID] = f
	case EventRemediationReviewed:
		var d RemediationReviewedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		f := c.Findings[d.FindingID]
		f.ReviewedBy, f.ReviewDecision, f.ReviewedAt = d.ReviewedBy, d.Decision, &d.ReviewedAt
		if d.Decision == DecisionAccept {
			f.Status = FindingAccepted
		} else {
			f.Status = FindingRejected
		}
		c.Findings[d.FindingID] = f
		if d.AllBlockersClosed {
			c.Status = StatusSiteVerification
		}
	case EventSiteVerificationPassed:
		var d SiteVerificationPassedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		c.SiteVerification, c.FrozenPlanRevision, c.SiteChecklistDigest = &d.Verification, d.PlanRevision, d.ChecklistDigest
		plan := c.Plans[d.PlanRevision]
		frozen := event.OccurredAt
		plan.FrozenAt = &frozen
		c.Plans[d.PlanRevision] = plan
		c.Status = StatusFrozen
	case EventClearanceCredentialIssued:
		if c.Credential != nil {
			return Conflict("放行凭据不可替换或重复签发")
		}
		var d ClearanceCredentialIssuedData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return err
		}
		credential := d.Credential
		c.Credential, c.Status = &credential, StatusCleared
	default:
		return fmt.Errorf("未知领域事件 %q", event.Type)
	}
	c.Version++
	c.UpdatedAt = event.OccurredAt
	return nil
}

func Replay(events []Event) (*RelocationCase, error) {
	c := NewCaseState()
	for _, event := range events {
		if err := c.Apply(event); err != nil {
			return nil, err
		}
	}
	return c, nil
}
