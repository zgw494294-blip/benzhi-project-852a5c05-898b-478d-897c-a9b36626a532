package domain

import (
	"encoding/json"
	"time"
)

const (
	EventCaseCreated               = "CaseCreated"
	EventRootSurveyRecorded        = "RootSurveyRecorded"
	EventProtectionPlanRevised     = "ProtectionPlanRevised"
	EventRiskReviewCompleted       = "RiskReviewCompleted"
	EventRemediationSubmitted      = "RemediationSubmitted"
	EventRemediationReviewed       = "RemediationReviewed"
	EventSiteVerificationPassed    = "SiteVerificationPassed"
	EventClearanceCredentialIssued = "ClearanceCredentialIssued"
)

type Event struct {
	Type       string          `json:"type"`
	CaseID     string          `json:"caseID"`
	OccurredAt time.Time       `json:"occurredAt"`
	Actor      string          `json:"actor"`
	Data       json.RawMessage `json:"data"`
}

func NewEvent(eventType, caseID, actor string, at time.Time, value any) (Event, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return Event{}, err
	}
	return Event{Type: eventType, CaseID: caseID, Actor: actor, OccurredAt: at.UTC(), Data: b}, nil
}

type CaseCreatedData struct {
	TreeCode           string          `json:"treeCode"`
	ProtectionGrade    ProtectionGrade `json:"protectionGrade"`
	Species            string          `json:"species"`
	TrunkDiameterCM    float64         `json:"trunkDiameterCM"`
	CrownRadiusM       float64         `json:"crownRadiusM"`
	PlannedWindowStart time.Time       `json:"plannedWindowStart"`
	PlannedWindowEnd   time.Time       `json:"plannedWindowEnd"`
}

type RootSurveyRecordedData struct {
	Survey               RootSurvey `json:"survey"`
	CoverageComplete     bool       `json:"coverageComplete"`
	EvidenceCompleteness float64    `json:"evidenceCompleteness"`
}

type ProtectionPlanRevisedData struct {
	Plan ProtectionPlan `json:"plan"`
}

type RiskReviewCompletedData struct {
	RuleVersion string        `json:"ruleVersion"`
	InputDigest string        `json:"inputDigest"`
	Findings    []RiskFinding `json:"findings"`
}

type RemediationSubmittedData struct {
	FindingID   string   `json:"findingID"`
	Evidence    []string `json:"evidence"`
	SubmittedBy string   `json:"submittedBy"`
}

type RemediationReviewedData struct {
	FindingID         string         `json:"findingID"`
	ReviewedBy        string         `json:"reviewedBy"`
	Decision          ReviewDecision `json:"decision"`
	ReviewedAt        time.Time      `json:"reviewedAt"`
	AllBlockersClosed bool           `json:"allBlockersClosed"`
}

type SiteVerificationPassedData struct {
	Verification    SiteVerification `json:"verification"`
	PlanRevision    int              `json:"planRevision"`
	ChecklistDigest string           `json:"checklistDigest"`
}

type ClearanceCredentialIssuedData struct {
	Credential ClearanceCredential `json:"credential"`
}
