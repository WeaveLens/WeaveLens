package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/elip/WeaveLens/internal/application/service"
	"github.com/elip/WeaveLens/internal/transport/security"
)

type ScanResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Region    string `json:"region"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type GraphResponse struct {
	Nodes []service.Resource     `json:"nodes"`
	Edges []service.Relationship `json:"edges"`
}

type ResourceResponse struct {
	service.Resource
}

type RelationshipResponse struct {
	service.Relationship
}

type HealthResponse struct {
	Status string `json:"status"`
}

type ReadyResponse struct {
	Status string `json:"status"`
}

type ConnectionStatusGetter interface {
	GetConnectionStatus() ConnectionStatus
}

type ConnectionStatus struct {
	State            string `json:"state"`
	AccountID        string `json:"accountId"`
	ARN              string `json:"arn"`
	Region           string `json:"region"`
	CredentialSource string `json:"credentialSource"`
	Message          string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func NewRouter(discovery service.DiscoveryService, graph service.GraphService, connection ConnectionStatusGetter, export service.ExportService, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HealthResponse{Status: "healthy"})
	})

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ReadyResponse{Status: "ready"})
	})

	mux.HandleFunc("GET /api/connection", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, connection.GetConnectionStatus())
	})

	mux.HandleFunc("POST /api/scans", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Region string `json:"region"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		scanID, err := discovery.StartScan(r.Context(), req.Region)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to start scan")
			logger.Warn("scan_start_failed", "error", err, "request_id", security.GetRequestID(r.Context()))
			return
		}

		status, _, _ := discovery.GetScanStatus(r.Context(), scanID)

		logger.Info("scan_started",
			"scan_id", scanID,
			"region", req.Region,
			"request_id", security.GetRequestID(r.Context()),
		)

		writeJSON(w, http.StatusAccepted, ScanResponse{
			ID:        scanID,
			Status:    status,
			Region:    req.Region,
			CreatedAt: parseTime(time.Now()),
			UpdatedAt: parseTime(time.Now()),
		})
	})

	mux.HandleFunc("GET /api/scans/{scanId}/status", func(w http.ResponseWriter, r *http.Request) {
		scanID := r.PathValue("scanId")
		status, _, err := discovery.GetScanStatus(r.Context(), scanID)
		if err != nil {
			writeError(w, http.StatusNotFound, "Scan not found")
			return
		}

		writeJSON(w, http.StatusOK, ScanResponse{
			ID:        scanID,
			Status:    status,
			UpdatedAt: parseTime(time.Now()),
		})
	})

	mux.HandleFunc("GET /api/scans/{scanId}/graph", func(w http.ResponseWriter, r *http.Request) {
		scanID := r.PathValue("scanId")
		nodes, edges, err := graph.GetGraph(r.Context(), scanID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get graph")
			return
		}

		writeJSON(w, http.StatusOK, GraphResponse{
			Nodes: nodes,
			Edges: edges,
		})
	})

	mux.HandleFunc("GET /api/resources/{resourceId}", func(w http.ResponseWriter, r *http.Request) {
		resourceID := r.PathValue("resourceId")
		resource, err := graph.GetResource(r.Context(), resourceID)
		if err != nil {
			writeError(w, http.StatusNotFound, "Resource not found")
			return
		}

		writeJSON(w, http.StatusOK, ResourceResponse{Resource: resource})
	})

	mux.HandleFunc("GET /api/resources/{resourceId}/relationships", func(w http.ResponseWriter, r *http.Request) {
		resourceID := r.PathValue("resourceId")
		relationships, err := graph.GetRelationships(r.Context(), resourceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get relationships")
			return
		}

		writeJSON(w, http.StatusOK, relationships)
	})

	mux.HandleFunc("GET /api/scans", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []ScanResponse{})
	})

	mux.HandleFunc("GET /api/scans/{scanId}/export", func(w http.ResponseWriter, r *http.Request) {
		scanID := r.PathValue("scanId")
		format := service.ExportFormat(r.URL.Query().Get("format"))
		if format == "" {
			format = service.ExportFormatJSON
		}

		data, err := export.ExportGraph(r.Context(), scanID, format)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to export graph")
			logger.Warn("export_failed",
				"scan_id", scanID,
				"format", format,
				"error", err,
				"request_id", security.GetRequestID(r.Context()),
			)
			return
		}

		logger.Info("graph_exported",
			"scan_id", scanID,
			"format", format,
			"request_id", security.GetRequestID(r.Context()),
		)

		switch format {
		case service.ExportFormatJSON:
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", "attachment; filename=\"graph.json\"")
		case service.ExportFormatDrawIO:
			w.Header().Set("Content-Type", "application/xml")
			w.Header().Set("Content-Disposition", "attachment; filename=\"graph.drawio\"")
		case service.ExportFormatSVG:
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Content-Disposition", "attachment; filename=\"graph.svg\"")
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	return mux
}

func StartServer(addr string, mux *http.ServeMux, apiKey string, logger *slog.Logger) (*http.Server, error) {
	var handler http.Handler = mux
	handler = security.RequestID(handler)
	handler = security.SecurityHeaders(handler)

	if apiKey != "" {
		handler = security.RequireAPIKey(apiKey, logger)(handler)
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return server, nil
}
