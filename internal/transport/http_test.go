package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elip/WeaveLens/internal/application/service"
	"github.com/elip/WeaveLens/internal/transport"
)

type fakeDiscoveryService struct {
	service.DiscoveryService
}

func (f *fakeDiscoveryService) StartScan(ctx context.Context, region string) (string, error) {
	return "scan-123", nil
}

func (f *fakeDiscoveryService) GetScanStatus(ctx context.Context, scanID string) (string, int, error) {
	return "RUNNING", 0, nil
}

func (f *fakeDiscoveryService) CancelScan(ctx context.Context, scanID string) error {
	return nil
}

func (f *fakeDiscoveryService) ListResources(ctx context.Context, scanID, category, resourceType string) ([]service.Resource, error) {
	return nil, nil
}

type fakeGraphService struct {
	service.GraphService
}

func (f *fakeGraphService) GetGraph(ctx context.Context, scanID string) ([]service.Resource, []service.Relationship, error) {
	return nil, nil, nil
}

func (f *fakeGraphService) GetResource(ctx context.Context, resourceID string) (service.Resource, error) {
	return service.Resource{}, nil
}

func (f *fakeGraphService) GetNeighbors(ctx context.Context, resourceID string) ([]service.Resource, error) {
	return nil, nil
}

func (f *fakeGraphService) GetRelationships(ctx context.Context, resourceID string) ([]service.Relationship, error) {
	return nil, nil
}

type fakeConnectionStatus struct{}

func (f *fakeConnectionStatus) GetConnectionStatus() transport.ConnectionStatus {
	return transport.ConnectionStatus{
		State:            "connected",
		AccountID:        "123456789012",
		ARN:              "arn:aws:iam::123456789012:role/test",
		Region:           "us-east-1",
		CredentialSource: "profile",
		Message:          "",
	}
}

func TestHealthEndpoint(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestReadyEndpoint(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/ready")
	if err != nil {
		t.Fatalf("GET /ready error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /ready status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestStartScan(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/scans", "application/json", strings.NewReader(`{"region":"us-east-1"}`))
	if err != nil {
		t.Fatalf("POST /api/scans error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("POST /api/scans status = %v, want %v", resp.StatusCode, http.StatusAccepted)
	}
}

func TestGetScanStatus(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/scans/scan-123/status")
	if err != nil {
		t.Fatalf("GET /api/scans/scan-123/status error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/scans/scan-123/status status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestGetGraph(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/scans/scan-123/graph")
	if err != nil {
		t.Fatalf("GET /api/scans/scan-123/graph error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/scans/scan-123/graph status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestGetResource(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/resources/res-1")
	if err != nil {
		t.Fatalf("GET /api/resources/res-1 error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/resources/res-1 status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestGetRelationships(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/resources/res-1/relationships")
	if err != nil {
		t.Fatalf("GET /api/resources/res-1/relationships error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/resources/res-1/relationships status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestScanStatusTimestamp(t *testing.T) {
	mux := transport.NewRouter(&fakeDiscoveryService{}, &fakeGraphService{}, &fakeConnectionStatus{})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/scans/scan-123/status")
	if err != nil {
		t.Fatalf("GET /api/scans/scan-123/status error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/scans/scan-123/status status = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	var scan transport.ScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&scan); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, err := time.Parse(time.RFC3339, scan.UpdatedAt); err != nil {
		t.Errorf("UpdatedAt is not RFC3339: %v", err)
	}
}
