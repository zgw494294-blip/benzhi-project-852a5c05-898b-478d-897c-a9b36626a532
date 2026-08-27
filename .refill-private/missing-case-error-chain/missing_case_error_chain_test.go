package missing_case_error_chain_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/eventstore"
	"heritage-tree-relocation-clearance/internal/httpapi"
)

func TestMissingCaseMutationPreservesNotFoundChain(t *testing.T) {
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	handler := httpapi.New(
		application.NewService(store),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	).Handler()

	tests := []struct {
		name string
		path string
		body string
	}{
		{
			name: "root survey",
			path: "/api/v1/relocation-cases/CASE-MISSING/root-surveys",
			body: `{"expectedVersion":0,"idempotencyKey":"missing-survey-key","sector":"N"}`,
		},
		{
			name: "protection plan",
			path: "/api/v1/relocation-cases/CASE-MISSING/protection-plans",
			body: `{"expectedVersion":0,"idempotencyKey":"missing-plan-key"}`,
		},
		{
			name: "risk review",
			path: "/api/v1/relocation-cases/CASE-MISSING/risk-reviews",
			body: `{"expectedVersion":0,"idempotencyKey":"missing-risk-key"}`,
		},
		{
			name: "remediation submission",
			path: "/api/v1/relocation-cases/CASE-MISSING/findings/FINDING-1/remediations",
			body: `{"expectedVersion":0,"idempotencyKey":"missing-remediation-key"}`,
		},
		{
			name: "remediation review",
			path: "/api/v1/relocation-cases/CASE-MISSING/findings/FINDING-1/reviews",
			body: `{"expectedVersion":0,"idempotencyKey":"missing-review-key"}`,
		},
		{
			name: "site verification",
			path: "/api/v1/relocation-cases/CASE-MISSING/site-verifications",
			body: `{"expectedVersion":0,"idempotencyKey":"missing-site-key"}`,
		},
		{
			name: "credential issuance",
			path: "/api/v1/relocation-cases/CASE-MISSING/credentials",
			body: `{"expectedVersion":0,"idempotencyKey":"missing-credential-key"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusNotFound {
				t.Fatalf("status=%d，期望 404，body=%s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"code":"NOT_FOUND"`) {
				t.Fatalf("响应未保留 NOT_FOUND：%s", response.Body.String())
			}
		})
	}
}
