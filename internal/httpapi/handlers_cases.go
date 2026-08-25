package httpapi

import (
	"net/http"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
)

type createCaseRequest struct {
	CaseID             string                 `json:"caseID,omitempty"`
	TreeCode           string                 `json:"treeCode"`
	ProtectionGrade    domain.ProtectionGrade `json:"protectionGrade"`
	Species            string                 `json:"species"`
	TrunkDiameterCM    float64                `json:"trunkDiameterCM"`
	CrownRadiusM       float64                `json:"crownRadiusM"`
	PlannedWindowStart time.Time              `json:"plannedWindowStart"`
	PlannedWindowEnd   time.Time              `json:"plannedWindowEnd"`
	CreatedBy          string                 `json:"createdBy"`
	IdempotencyKey     string                 `json:"idempotencyKey"`
}

func (a *API) CreateCase(w http.ResponseWriter, r *http.Request) {
	var input createCaseRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	result, err := a.service.CreateCase(application.CreateCaseCommand{CaseID: input.CaseID, TreeCode: input.TreeCode, ProtectionGrade: input.ProtectionGrade, Species: input.Species, TrunkDiameterCM: input.TrunkDiameterCM, CrownRadiusM: input.CrownRadiusM, PlannedWindowStart: input.PlannedWindowStart, PlannedWindowEnd: input.PlannedWindowEnd, CreatedBy: input.CreatedBy, IdempotencyKey: input.IdempotencyKey})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	w.Header().Set("Location", "/api/v1/relocation-cases/"+result.CaseID)
	writeJSON(w, http.StatusCreated, result)
}

func (a *API) GetCase(w http.ResponseWriter, r *http.Request) {
	caseState, err := a.service.GetCase(r.PathValue("caseID"))
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, caseState)
}

type surveyRequest struct {
	writeMeta
	SurveyID            string   `json:"surveyID,omitempty"`
	Sector              string   `json:"sector"`
	ProbeDepthCM        float64  `json:"probeDepthCM"`
	CriticalRootCount   int      `json:"criticalRootCount"`
	ExposedRootRatio    float64  `json:"exposedRootRatio"`
	SoilMoisturePercent float64  `json:"soilMoisturePercent"`
	EvidenceRefs        []string `json:"evidenceRefs"`
	RecordedBy          string   `json:"recordedBy"`
}

func (a *API) RecordRootSurvey(w http.ResponseWriter, r *http.Request) {
	var input surveyRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	result, err := a.service.RecordSurvey(application.RecordSurveyCommand{CaseID: r.PathValue("caseID"), WriteContext: application.WriteContext{ExpectedVersion: input.ExpectedVersion, IdempotencyKey: input.IdempotencyKey}, SurveyID: input.SurveyID, Sector: input.Sector, ProbeDepthCM: input.ProbeDepthCM, CriticalRootCount: input.CriticalRootCount, ExposedRootRatio: input.ExposedRootRatio, SoilMoisturePercent: input.SoilMoisturePercent, EvidenceRefs: input.EvidenceRefs, RecordedBy: input.RecordedBy})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
