package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	maxSavedScans = 20
	historyFile   = ".scans.json"
)

type ScanHistoryEntry struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Region    string    `json:"region"`
	NodeCount int       `json:"nodeCount"`
	EdgeCount int       `json:"edgeCount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type GraphData struct {
	Nodes []Resource     `json:"nodes"`
	Edges []Relationship `json:"edges"`
}

type ScanHistoryData struct {
	Scans  []ScanHistoryEntry   `json:"scans"`
	Graphs map[string]GraphData `json:"graphs"`
}

type ScanHistory struct {
	mu       sync.RWMutex
	data     ScanHistoryData
	filePath string
}

func NewScanHistory() *ScanHistory {
	h := &ScanHistory{
		data: ScanHistoryData{
			Scans:  []ScanHistoryEntry{},
			Graphs: map[string]GraphData{},
		},
		filePath: filepath.Join(".", historyFile),
	}
	h.load()
	return h
}

func (h *ScanHistory) AddScan(scanID, region string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().UTC()
	entry := ScanHistoryEntry{
		ID:        scanID,
		Status:    "RUNNING",
		Region:    region,
		CreatedAt: now,
		UpdatedAt: now,
	}
	h.data.Scans = append([]ScanHistoryEntry{entry}, h.data.Scans...)
	h.truncate()
	h.save()
}

func (h *ScanHistory) UpdateScan(scanID, status string, nodeCount, edgeCount int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().UTC()
	for i, scan := range h.data.Scans {
		if scan.ID == scanID {
			h.data.Scans[i].Status = status
			h.data.Scans[i].NodeCount = nodeCount
			h.data.Scans[i].EdgeCount = edgeCount
			h.data.Scans[i].UpdatedAt = now
			break
		}
	}
	h.save()
}

func (h *ScanHistory) SaveGraph(scanID string, nodes []Resource, edges []Relationship) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data.Graphs[scanID] = GraphData{
		Nodes: nodes,
		Edges: edges,
	}
	h.save()
}

func (h *ScanHistory) GetScans() []ScanHistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]ScanHistoryEntry, len(h.data.Scans))
	copy(result, h.data.Scans)
	return result
}

func (h *ScanHistory) GetGraph(scanID string) (GraphData, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	graph, exists := h.data.Graphs[scanID]
	return graph, exists
}

func (h *ScanHistory) truncate() {
	if len(h.data.Scans) > maxSavedScans {
		removed := h.data.Scans[maxSavedScans:]
		h.data.Scans = h.data.Scans[:maxSavedScans]
		for _, scan := range removed {
			delete(h.data.Graphs, scan.ID)
		}
	}
}

func (h *ScanHistory) load() {
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return
	}
	var history ScanHistoryData
	if err := json.Unmarshal(data, &history); err != nil {
		return
	}
	h.data = history
}

func (h *ScanHistory) save() {
	data, err := json.MarshalIndent(h.data, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(h.filePath, data, 0644)
}

func (h *ScanHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data = ScanHistoryData{
		Scans:  []ScanHistoryEntry{},
		Graphs: map[string]GraphData{},
	}
	h.save()
}
