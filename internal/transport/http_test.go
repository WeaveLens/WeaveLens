package transport_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elip/WeaveLens/internal/transport"
)

func TestHealthEndpoint(t *testing.T) {
	mux := transport.NewRouter()
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
	mux := transport.NewRouter()
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
