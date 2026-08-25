package httpapi

import (
	"net/http"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
)

type remediationRequest struct {
	writeMeta
	Evidence    []string `json:"evidence"`
	SubmittedBy string   `json:"submittedBy"`
}

func (a *API) SubmitRemediation(w http.ResponseWriter, r *http.Request) {
	var input remediationRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	result, err := a.service.SubmitRemediation(application.SubmitRemediationCommand{CaseID: r.PathValue("caseID"), WriteContext: application.WriteContext{ExpectedVersion: input.ExpectedVersion, IdempotencyKey: input.IdempotencyKey}, FindingID: r.PathValue("findingID"), Evidence: input.Evidence, SubmittedBy: input.SubmittedBy})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type remediationReviewRequest struct {
	writeMeta
	ReviewedBy string                `json:"reviewedBy"`
	Decision   domain.ReviewDecision `json:"decision"`
}

func (a *API) ReviewRemediation(w http.ResponseWriter, r *http.Request) {
	var input remediationReviewRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	result, err := a.service.ReviewRemediation(application.ReviewRemediationCommand{CaseID: r.PathValue("caseID"), WriteContext: application.WriteContext{ExpectedVersion: input.ExpectedVersion, IdempotencyKey: input.IdempotencyKey}, FindingID: r.PathValue("findingID"), ReviewedBy: input.ReviewedBy, Decision: input.Decision})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
