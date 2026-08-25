package httpapi

import (
	"net/http"

	"heritage-tree-relocation-clearance/internal/application"
)

type planRequest struct {
	writeMeta
	PlanID               string             `json:"planID,omitempty"`
	CutBoundaryCM        float64            `json:"cutBoundaryCM"`
	RootBallRadiusCM     float64            `json:"rootBallRadiusCM"`
	RootBallDepthCM      float64            `json:"rootBallDepthCM"`
	SupportMethod        string             `json:"supportMethod"`
	MoistureMethod       string             `json:"moistureMethod"`
	MaxTransportMinutes  int                `json:"maxTransportMinutes"`
	MonitoringThresholds map[string]float64 `json:"monitoringThresholds"`
	PreparedBy           string             `json:"preparedBy"`
}

func (a *API) ReviseProtectionPlan(w http.ResponseWriter, r *http.Request) {
	var input planRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	result, err := a.service.RevisePlan(application.RevisePlanCommand{CaseID: r.PathValue("caseID"), WriteContext: application.WriteContext{ExpectedVersion: input.ExpectedVersion, IdempotencyKey: input.IdempotencyKey}, PlanID: input.PlanID, CutBoundaryCM: input.CutBoundaryCM, RootBallRadiusCM: input.RootBallRadiusCM, RootBallDepthCM: input.RootBallDepthCM, SupportMethod: input.SupportMethod, MoistureMethod: input.MoistureMethod, MaxTransportMinutes: input.MaxTransportMinutes, MonitoringThresholds: input.MonitoringThresholds, PreparedBy: input.PreparedBy})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

type riskReviewRequest struct {
	writeMeta
	ReviewedBy string `json:"reviewedBy"`
}

func (a *API) ReviewRisk(w http.ResponseWriter, r *http.Request) {
	var input riskReviewRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	result, err := a.service.ReviewRisk(application.ReviewRiskCommand{CaseID: r.PathValue("caseID"), WriteContext: application.WriteContext{ExpectedVersion: input.ExpectedVersion, IdempotencyKey: input.IdempotencyKey}, ReviewedBy: input.ReviewedBy})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
