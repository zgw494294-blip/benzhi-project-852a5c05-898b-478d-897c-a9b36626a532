package domain

import "time"

type CaseStatus string

const (
	StatusPreparing        CaseStatus = "PREPARING"
	StatusPlanning         CaseStatus = "PLANNING"
	StatusRemediation      CaseStatus = "REMEDIATION"
	StatusSiteVerification CaseStatus = "SITE_VERIFICATION"
	StatusFrozen           CaseStatus = "FROZEN"
	StatusCleared          CaseStatus = "CLEARED"
)

type ProtectionGrade string

const (
	GradeOne   ProtectionGrade = "I"
	GradeTwo   ProtectionGrade = "II"
	GradeThree ProtectionGrade = "III"
)

type Severity string

const (
	SeverityBlocker Severity = "BLOCKER"
	SeverityWarning Severity = "WARNING"
)

type FindingStatus string

const (
	FindingOpen      FindingStatus = "OPEN"
	FindingSubmitted FindingStatus = "SUBMITTED"
	FindingAccepted  FindingStatus = "ACCEPTED"
	FindingRejected  FindingStatus = "REJECTED"
)

type ReviewDecision string

const (
	DecisionAccept ReviewDecision = "ACCEPT"
	DecisionReject ReviewDecision = "REJECT"
)

type RelocationCase struct {
	CaseID              string                 `json:"caseID"`
	TreeCode            string                 `json:"treeCode"`
	ProtectionGrade     ProtectionGrade        `json:"protectionGrade"`
	Species             string                 `json:"species"`
	TrunkDiameterCM     float64                `json:"trunkDiameterCM"`
	CrownRadiusM        float64                `json:"crownRadiusM"`
	PlannedWindowStart  time.Time              `json:"plannedWindowStart"`
	PlannedWindowEnd    time.Time              `json:"plannedWindowEnd"`
	Status              CaseStatus             `json:"status"`
	Version             uint64                 `json:"version"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
	Surveys             map[string]RootSurvey  `json:"surveys"`
	Plans               map[int]ProtectionPlan `json:"plans"`
	CurrentPlanRevision int                    `json:"currentPlanRevision"`
	Findings            map[string]RiskFinding `json:"findings"`
	ReviewRuleVersion   string                 `json:"reviewRuleVersion,omitempty"`
	ReviewInputDigest   string                 `json:"reviewInputDigest,omitempty"`
	SiteVerification    *SiteVerification      `json:"siteVerification,omitempty"`
	Credential          *ClearanceCredential   `json:"credential,omitempty"`
	FrozenPlanRevision  int                    `json:"frozenPlanRevision,omitempty"`
	SiteChecklistDigest string                 `json:"siteChecklistDigest,omitempty"`
}

type RootSurvey struct {
	SurveyID                    string    `json:"surveyID"`
	CaseID                      string    `json:"caseID"`
	Sector                      string    `json:"sector"`
	ProbeDepthCM                float64   `json:"probeDepthCM"`
	CriticalRootCount           int       `json:"criticalRootCount"`
	ExposedRootRatio            float64   `json:"exposedRootRatio"`
	SoilMoisturePercent         float64   `json:"soilMoisturePercent"`
	EvidenceRefs                []string  `json:"evidenceRefs"`
	RecommendedRootBallRadiusCM float64   `json:"recommendedRootBallRadiusCM"`
	RecordedBy                  string    `json:"recordedBy"`
	RecordedAt                  time.Time `json:"recordedAt"`
}

type ProtectionPlan struct {
	PlanID               string             `json:"planID"`
	CaseID               string             `json:"caseID"`
	Revision             int                `json:"revision"`
	CutBoundaryCM        float64            `json:"cutBoundaryCM"`
	RootBallRadiusCM     float64            `json:"rootBallRadiusCM"`
	RootBallDepthCM      float64            `json:"rootBallDepthCM"`
	SupportMethod        string             `json:"supportMethod"`
	MoistureMethod       string             `json:"moistureMethod"`
	MaxTransportMinutes  int                `json:"maxTransportMinutes"`
	MonitoringThresholds map[string]float64 `json:"monitoringThresholds"`
	InputDigest          string             `json:"inputDigest"`
	FrozenAt             *time.Time         `json:"frozenAt,omitempty"`
	PreparedBy           string             `json:"preparedBy"`
	PreparedAt           time.Time          `json:"preparedAt"`
}

type RiskFinding struct {
	FindingID           string         `json:"findingID"`
	CaseID              string         `json:"caseID"`
	RuleCode            string         `json:"ruleCode"`
	Severity            Severity       `json:"severity"`
	Description         string         `json:"description"`
	Status              FindingStatus  `json:"status"`
	RemediationEvidence []string       `json:"remediationEvidence,omitempty"`
	SubmittedBy         string         `json:"submittedBy,omitempty"`
	ReviewedBy          string         `json:"reviewedBy,omitempty"`
	ReviewDecision      ReviewDecision `json:"reviewDecision,omitempty"`
	ReviewedAt          *time.Time     `json:"reviewedAt,omitempty"`
}

type SiteVerification struct {
	WorkZoneReady        bool      `json:"workZoneReady"`
	MachineryAccessReady bool      `json:"machineryAccessReady"`
	TemporaryCareReady   bool      `json:"temporaryCareReady"`
	WeatherWindowSafe    bool      `json:"weatherWindowSafe"`
	Notes                string    `json:"notes"`
	VerifiedBy           string    `json:"verifiedBy"`
	VerifiedAt           time.Time `json:"verifiedAt"`
}

type ClearanceCredential struct {
	CredentialID        string    `json:"credentialID"`
	CaseID              string    `json:"caseID"`
	SerialNumber        uint64    `json:"serialNumber"`
	FrozenPlanRevision  int       `json:"frozenPlanRevision"`
	SiteChecklistDigest string    `json:"siteChecklistDigest"`
	EventSequence       uint64    `json:"eventSequence"`
	ContentDigest       string    `json:"contentDigest"`
	IssuedBy            string    `json:"issuedBy"`
	IssuedAt            time.Time `json:"issuedAt"`
}

func NewCaseState() *RelocationCase {
	return &RelocationCase{Surveys: map[string]RootSurvey{}, Plans: map[int]ProtectionPlan{}, Findings: map[string]RiskFinding{}}
}
