package canceledrequestcommit_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/eventstore"
	"heritage-tree-relocation-clearance/internal/httpapi"
)

func TestCanceledRequestDoesNotCommit(t *testing.T) {
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	body := []byte(`{"caseID":"CASE-CANCELED","treeCode":"TREE-CANCEL","protectionGrade":"I","species":"香樟","trunkDiameterCM":80,"crownRadiusM":6,"plannedWindowStart":"2099-09-01T00:00:00Z","plannedWindowEnd":"2099-09-02T00:00:00Z","createdBy":"档案员甲","idempotencyKey":"canceled-create-key"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/relocation-cases", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	ctx, cancel := context.WithCancel(request.Context())
	cancel()
	request = request.WithContext(ctx)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	httpapi.New(application.NewService(store), logger).Handler().ServeHTTP(httptest.NewRecorder(), request)
	if store.CaseExists("CASE-CANCELED") || store.LastSequence() != 0 {
		t.Fatalf("canceled request committed case: exists=%t sequence=%d", store.CaseExists("CASE-CANCELED"), store.LastSequence())
	}
}
