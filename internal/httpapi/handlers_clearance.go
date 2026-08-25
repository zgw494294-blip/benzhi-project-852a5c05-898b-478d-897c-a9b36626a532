package httpapi

import (
	"net/http"

	"heritage-tree-relocation-clearance/internal/application"
)

type siteVerificationRequest struct {
	writeMeta
	WorkZoneReady        bool   `json:"workZoneReady"`
	MachineryAccessReady bool   `json:"machineryAccessReady"`
	TemporaryCareReady   bool   `json:"temporaryCareReady"`
	WeatherWindowSafe    bool   `json:"weatherWindowSafe"`
	Notes                string `json:"notes"`
	VerifiedBy           string `json:"verifiedBy"`
}

func (a *API) VerifySite(w http.ResponseWriter, r *http.Request) {
	var input siteVerificationRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	result, err := a.service.VerifySite(application.VerifySiteCommand{CaseID: r.PathValue("caseID"), WriteContext: application.WriteContext{ExpectedVersion: input.ExpectedVersion, IdempotencyKey: input.IdempotencyKey}, WorkZoneReady: input.WorkZoneReady, MachineryAccessReady: input.MachineryAccessReady, TemporaryCareReady: input.TemporaryCareReady, WeatherWindowSafe: input.WeatherWindowSafe, Notes: input.Notes, VerifiedBy: input.VerifiedBy})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type credentialRequest struct {
	writeMeta
	IssuedBy string `json:"issuedBy"`
}

func (a *API) IssueCredential(w http.ResponseWriter, r *http.Request) {
	var input credentialRequest
	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	credential, result, err := a.service.IssueCredential(application.IssueCredentialCommand{CaseID: r.PathValue("caseID"), WriteContext: application.WriteContext{ExpectedVersion: input.ExpectedVersion, IdempotencyKey: input.IdempotencyKey}, IssuedBy: input.IssuedBy})
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	w.Header().Set("Location", "/api/v1/clearance-credentials/"+credential.CredentialID)
	writeJSON(w, http.StatusCreated, map[string]any{"credential": credential, "mutation": result})
}

func (a *API) GetCredential(w http.ResponseWriter, r *http.Request) {
	credential, err := a.service.GetCredential(r.PathValue("credentialID"))
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, credential)
}

func (a *API) GetAudit(w http.ResponseWriter, r *http.Request) {
	audit, err := a.service.GetAudit(r.PathValue("caseID"))
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, audit)
}

func (a *API) GetCredentialAudit(w http.ResponseWriter, r *http.Request) {
	audit, err := a.service.GetAuditByCredential(r.PathValue("credentialID"))
	if err != nil {
		respondError(w, r, a.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, audit)
}
