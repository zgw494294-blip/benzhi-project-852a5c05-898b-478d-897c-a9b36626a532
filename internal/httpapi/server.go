package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
)

type contextKey string

const requestIDKey contextKey = "requestID"

type API struct {
	service *application.Service
	logger  *slog.Logger
	mux     *http.ServeMux
}

func New(service *application.Service, logger *slog.Logger) *API {
	if logger == nil {
		logger = slog.Default()
	}
	api := &API{service: service, logger: logger, mux: http.NewServeMux()}
	api.routes()
	return api
}

func (a *API) Handler() http.Handler { return a.middleware(a.mux) }

func (a *API) routes() {
	a.mux.HandleFunc("GET /healthz", a.Health)
	a.mux.HandleFunc("POST /api/v1/relocation-cases", a.CreateCase)
	a.mux.HandleFunc("GET /api/v1/relocation-cases/{caseID}", a.GetCase)
	a.mux.HandleFunc("POST /api/v1/relocation-cases/{caseID}/root-surveys", a.RecordRootSurvey)
	a.mux.HandleFunc("POST /api/v1/relocation-cases/{caseID}/protection-plans", a.ReviseProtectionPlan)
	a.mux.HandleFunc("POST /api/v1/relocation-cases/{caseID}/risk-reviews", a.ReviewRisk)
	a.mux.HandleFunc("POST /api/v1/relocation-cases/{caseID}/findings/{findingID}/remediations", a.SubmitRemediation)
	a.mux.HandleFunc("POST /api/v1/relocation-cases/{caseID}/findings/{findingID}/reviews", a.ReviewRemediation)
	a.mux.HandleFunc("POST /api/v1/relocation-cases/{caseID}/site-verifications", a.VerifySite)
	a.mux.HandleFunc("POST /api/v1/relocation-cases/{caseID}/credentials", a.IssueCredential)
	a.mux.HandleFunc("GET /api/v1/relocation-cases/{caseID}/audit", a.GetAudit)
	a.mux.HandleFunc("GET /api/v1/clearance-credentials/{credentialID}", a.GetCredential)
	a.mux.HandleFunc("GET /api/v1/clearance-credentials/{credentialID}/audit", a.GetCredentialAudit)
}

func (a *API) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" || len(requestID) > 100 {
			requestID = newRequestID()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		defer func() {
			if recovered := recover(); recovered != nil {
				a.logger.Error("HTTP handler panic", "requestID", requestID, "panic", recovered)
				respondError(w, r.WithContext(ctx), a.logger, context.Canceled)
			}
			a.logger.Debug("HTTP 请求完成", "requestID", requestID, "method", r.Method, "path", r.URL.Path, "elapsed", time.Since(started))
		}()
		if err := ctx.Err(); err != nil {
			respondError(w, r.WithContext(ctx), a.logger, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestIDFrom(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func newRequestID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "req-fallback"
	}
	return "req-" + hex.EncodeToString(b)
}

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "requestID": requestIDFrom(r.Context())})
}
