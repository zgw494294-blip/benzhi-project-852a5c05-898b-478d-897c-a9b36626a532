package application

import (
	"time"

	"heritage-tree-relocation-clearance/internal/domain"
)

type WriteContext struct {
	ExpectedVersion uint64
	IdempotencyKey  string
}

type CreateCaseCommand struct {
	CaseID             string
	TreeCode           string
	ProtectionGrade    domain.ProtectionGrade
	Species            string
	TrunkDiameterCM    float64
	CrownRadiusM       float64
	PlannedWindowStart time.Time
	PlannedWindowEnd   time.Time
	CreatedBy          string
	IdempotencyKey     string
}

type RecordSurveyCommand struct {
	CaseID              string
	WriteContext        WriteContext
	SurveyID            string
	Sector              string
	ProbeDepthCM        float64
	CriticalRootCount   int
	ExposedRootRatio    float64
	SoilMoisturePercent float64
	EvidenceRefs        []string
	RecordedBy          string
}

type RevisePlanCommand struct {
	CaseID               string
	WriteContext         WriteContext
	PlanID               string
	CutBoundaryCM        float64
	RootBallRadiusCM     float64
	RootBallDepthCM      float64
	SupportMethod        string
	MoistureMethod       string
	MaxTransportMinutes  int
	MonitoringThresholds map[string]float64
	PreparedBy           string
}

type ReviewRiskCommand struct {
	CaseID       string
	WriteContext WriteContext
	ReviewedBy   string
}

type SubmitRemediationCommand struct {
	CaseID       string
	WriteContext WriteContext
	FindingID    string
	Evidence     []string
	SubmittedBy  string
}

type ReviewRemediationCommand struct {
	CaseID       string
	WriteContext WriteContext
	FindingID    string
	ReviewedBy   string
	Decision     domain.ReviewDecision
}

type VerifySiteCommand struct {
	CaseID               string
	WriteContext         WriteContext
	WorkZoneReady        bool
	MachineryAccessReady bool
	TemporaryCareReady   bool
	WeatherWindowSafe    bool
	Notes                string
	VerifiedBy           string
}

type IssueCredentialCommand struct {
	CaseID       string
	WriteContext WriteContext
	IssuedBy     string
}

type MutationResult struct {
	CaseID     string            `json:"caseID"`
	Version    uint64            `json:"version"`
	Sequence   uint64            `json:"sequence"`
	Status     domain.CaseStatus `json:"status"`
	Idempotent bool              `json:"idempotent"`
}
