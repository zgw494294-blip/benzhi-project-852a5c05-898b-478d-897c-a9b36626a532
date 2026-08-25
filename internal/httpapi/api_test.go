package httpapi

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/eventstore"
)

func TestRunSelfCheckAgainstRealHTTPServer(t *testing.T) {
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(New(application.NewService(store), logger).Handler())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := RunSelfCheck(ctx, server.URL); err != nil {
		t.Fatal(err)
	}
}

func TestStrictJSONAndContentType(t *testing.T) {
	store, err := eventstore.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	server := httptest.NewServer(New(application.NewService(store), slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	defer server.Close()
	body := `{"caseID":"CASE-HTTP","treeCode":"TREE-HTTP","protectionGrade":"I","species":"香樟","trunkDiameterCM":80,"crownRadiusM":5,"plannedWindowStart":"2026-09-01T00:00:00Z","plannedWindowEnd":"2026-09-02T00:00:00Z","createdBy":"档案员甲","idempotencyKey":"http-create-key","unknown":true}`
	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/relocation-cases", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(payload), "INVALID_JSON") {
		t.Fatalf("status=%d body=%s", response.StatusCode, payload)
	}
	request, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/relocation-cases", bytes.NewBufferString(`{}`))
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("status=%d", response.StatusCode)
	}
}
