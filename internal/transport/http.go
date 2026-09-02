package transport

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/elip/WeaveLens/internal/application/service"
	"github.com/elip/WeaveLens/internal/transport/security"
)

type ScanResponse struct {
	ID        string   `json:"id"`
	Status    string   `json:"status"`
	Region    string   `json:"region"`
	Regions   []string `json:"regions,omitempty"`
	NodeCount int      `json:"nodeCount"`
	EdgeCount int      `json:"edgeCount"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
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
	return t.UTC().Format(time.RFC3339)
}

func normalizeRegions(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, r := range in {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func NewRouter(discovery service.DiscoveryService, graph service.GraphService, connection ConnectionStatusGetter, export service.ExportService, regions *service.RegionService, logger *slog.Logger, notifiers ...<-chan struct{}) *http.ServeMux {
	mux := http.NewServeMux()
	broadcaster := NewScanBroadcaster()

	var scanNotifier <-chan struct{}
	if len(notifiers) > 0 {
		scanNotifier = notifiers[0]
	}
	if scanNotifier != nil {
		go func() {
			for range scanNotifier {
				broadcaster.Broadcast(discovery.GetScans())
			}
		}()
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HealthResponse{Status: "healthy"})
	})

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ReadyResponse{Status: "ready"})
	})

	mux.HandleFunc("GET /api/connection", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, connection.GetConnectionStatus())
	})

	mux.HandleFunc("GET /api/regions", func(w http.ResponseWriter, r *http.Request) {
		if regions == nil {
			writeJSON(w, http.StatusOK, []service.RegionInfo{})
			return
		}
		writeJSON(w, http.StatusOK, regions.GetRegions(r.Context()))
	})

	mux.HandleFunc("POST /api/scans", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Region  string   `json:"region"`
			Regions []string `json:"regions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		regions := normalizeRegions(req.Regions)
		if regions == nil && req.Region != "" {
			regions = []string{req.Region}
		}

		scanID, err := discovery.StartScan(r.Context(), regions)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to start scan")
			logger.Warn("scan_start_failed", "error", err, "request_id", security.GetRequestID(r.Context()))
			return
		}

		status, _, _ := discovery.GetScanStatus(r.Context(), scanID)

		displayRegion := "all"
		if len(regions) == 1 {
			displayRegion = regions[0]
		} else if len(regions) > 1 {
			displayRegion = strings.Join(regions, ",")
		}

		logger.Info("scan_started",
			"scan_id", scanID,
			"regions", regions,
			"request_id", security.GetRequestID(r.Context()),
		)

		writeJSON(w, http.StatusAccepted, ScanResponse{
			ID:        scanID,
			Status:    status,
			Region:    displayRegion,
			Regions:   regions,
			CreatedAt: parseTime(time.Now().UTC()),
			UpdatedAt: parseTime(time.Now().UTC()),
		})
	})

	mux.HandleFunc("GET /api/scans/{scanId}/status", func(w http.ResponseWriter, r *http.Request) {
		scanID := r.PathValue("scanId")
		status, _, err := discovery.GetScanStatus(r.Context(), scanID)
		if err != nil {
			writeError(w, http.StatusNotFound, "Scan not found")
			return
		}

		region := ""
		regions := []string{}
		nodeCount := 0
		edgeCount := 0
		createdAt := time.Now().UTC()
		updatedAt := time.Now().UTC()
		for _, scan := range discovery.GetScans() {
			if scan.ID == scanID {
				region = scan.Region
				regions = scan.Regions
				nodeCount = scan.NodeCount
				edgeCount = scan.EdgeCount
				if !scan.CreatedAt.IsZero() {
					createdAt = scan.CreatedAt.UTC()
				}
				if !scan.UpdatedAt.IsZero() {
					updatedAt = scan.UpdatedAt.UTC()
				}
				break
			}
		}

		writeJSON(w, http.StatusOK, ScanResponse{
			ID:        scanID,
			Status:    status,
			Region:    region,
			Regions:   regions,
			NodeCount: nodeCount,
			EdgeCount: edgeCount,
			CreatedAt: parseTime(createdAt),
			UpdatedAt: parseTime(updatedAt),
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
		scans := discovery.GetScans()
		writeJSON(w, http.StatusOK, scans)
	})

	mux.HandleFunc("GET /api/scans/stream", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := broadcaster.Subscribe()
		defer broadcaster.Unsubscribe(ch)

		scans := discovery.GetScans()
		if data, err := json.Marshal(scans); err == nil {
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				w.Write(msg)
				flusher.Flush()
			}
		}
	})

	mux.HandleFunc("DELETE /api/scans/{scanId}", func(w http.ResponseWriter, r *http.Request) {
		scanID := r.PathValue("scanId")
		deleted, err := discovery.DeleteScan(r.Context(), scanID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete scan")
			logger.Warn("scan_delete_failed",
				"scan_id", scanID,
				"error", err,
				"request_id", security.GetRequestID(r.Context()),
			)
			return
		}
		if !deleted {
			writeError(w, http.StatusConflict, "Scan not found or cannot be deleted while running")
			return
		}

		logger.Info("scan_deleted",
			"scan_id", scanID,
			"request_id", security.GetRequestID(r.Context()),
		)

		writeJSON(w, http.StatusOK, map[string]string{"id": scanID, "status": "deleted"})
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
		WriteTimeout: 0,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return server, nil
}
