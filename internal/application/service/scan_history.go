package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	maxSavedScans = 20
	historyFile   = "data/scans.json"
)

type ScanHistoryEntry struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Region    string    `json:"region"`
	Regions   []string  `json:"regions,omitempty"`
	NodeCount int       `json:"nodeCount"`
	EdgeCount int       `json:"edgeCount"`
	Pinned    bool      `json:"pinned,omitempty"`
	Locked    bool      `json:"locked,omitempty"`
	Layout    string    `json:"layout,omitempty"`
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
	notify   chan struct{}
}

func NewScanHistory() *ScanHistory {
	h := &ScanHistory{
		data: ScanHistoryData{
			Scans:  []ScanHistoryEntry{},
			Graphs: map[string]GraphData{},
		},
		filePath: resolveHistoryPath(),
		notify:   make(chan struct{}, 1),
	}
	h.load()
	return h
}

func resolveHistoryPath() string {
	if path := os.Getenv("WEAVELENS_HISTORY_FILE"); path != "" {
		return path
	}

	dir, err := os.Getwd()
	if err != nil {
		return filepath.Join(".", historyFile)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, historyFile)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join(".", historyFile)
		}
		dir = parent
	}
}

func (h *ScanHistory) Notify() <-chan struct{} {
	return h.notify
}

func (h *ScanHistory) signalChange() {
	select {
	case h.notify <- struct{}{}:
	default:
	}
}

func (h *ScanHistory) OnFileDeleted() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data = ScanHistoryData{
		Scans:  []ScanHistoryEntry{},
		Graphs: map[string]GraphData{},
	}
	h.signalChange()
}

func (h *ScanHistory) FilePath() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.filePath
}

func (h *ScanHistory) OnFileCreated() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.load()
	h.signalChange()
}

func (h *ScanHistory) AddScan(scanID, region string, regions []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().UTC()
	displayRegion := region
	if displayRegion == "" && len(regions) == 0 {
		displayRegion = "all"
	} else if len(regions) > 1 {
		displayRegion = strings.Join(regions, ",")
	}
	entry := ScanHistoryEntry{
		ID:        scanID,
		Status:    "RUNNING",
		Region:    displayRegion,
		Regions:   regions,
		CreatedAt: now,
		UpdatedAt: now,
	}
	h.data.Scans = append([]ScanHistoryEntry{entry}, h.data.Scans...)
	h.truncate()
	h.save()
	h.signalChange()
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
	h.signalChange()
}

func (h *ScanHistory) RemoveScan(scanID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, scan := range h.data.Scans {
		if scan.ID == scanID {
			h.data.Scans = append(h.data.Scans[:i], h.data.Scans[i+1:]...)
			delete(h.data.Graphs, scanID)
			h.save()
			h.signalChange()
			return true
		}
	}
	return false
}

func (h *ScanHistory) RemoveUnpinned() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	kept := h.data.Scans[:0]
	removed := 0
	for _, scan := range h.data.Scans {
		if scan.Pinned {
			kept = append(kept, scan)
			continue
		}
		delete(h.data.Graphs, scan.ID)
		removed++
	}
	h.data.Scans = append([]ScanHistoryEntry{}, kept...)
	if removed > 0 {
		h.save()
		h.signalChange()
	}
	return removed
}

func (h *ScanHistory) SetScanLocked(scanID string, locked bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, scan := range h.data.Scans {
		if scan.ID == scanID {
			if h.data.Scans[i].Locked == locked {
				return true
			}
			h.data.Scans[i].Locked = locked
			h.data.Scans[i].UpdatedAt = time.Now().UTC()
			h.save()
			h.signalChange()
			return true
		}
	}
	return false
}

func (h *ScanHistory) SetScanLayout(scanID string, layout string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, scan := range h.data.Scans {
		if scan.ID == scanID {
			if h.data.Scans[i].Layout == layout {
				return true
			}
			h.data.Scans[i].Layout = layout
			h.data.Scans[i].UpdatedAt = time.Now().UTC()
			h.save()
			h.signalChange()
			return true
		}
	}
	return false
}

func (h *ScanHistory) SetScanPinned(scanID string, pinned bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, scan := range h.data.Scans {
		if scan.ID == scanID {
			if h.data.Scans[i].Pinned == pinned {
				return true
			}
			h.data.Scans[i].Pinned = pinned
			h.data.Scans[i].UpdatedAt = time.Now().UTC()
			h.save()
			h.signalChange()
			return true
		}
	}
	return false
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
	h.mu.Lock()
	defer h.mu.Unlock()

	h.load()
	result := make([]ScanHistoryEntry, len(h.data.Scans))
	copy(result, h.data.Scans)
	return result
}

func (h *ScanHistory) FindScan(scanID string) (ScanHistoryEntry, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.load()
	for _, scan := range h.data.Scans {
		if scan.ID == scanID {
			return scan, true
		}
	}
	return ScanHistoryEntry{}, false
}

func (h *ScanHistory) GetGraph(scanID string) (GraphData, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.load()

	if h.data.Graphs == nil {
		h.data.Graphs = make(map[string]GraphData)
	}

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
		h.data = ScanHistoryData{
			Scans:  []ScanHistoryEntry{},
			Graphs: map[string]GraphData{},
		}
		return
	}
	var history ScanHistoryData
	if err := json.Unmarshal(data, &history); err != nil {
		return
	}
	h.data = history

	if h.data.Graphs == nil {
		h.data.Graphs = make(map[string]GraphData)
	}
}

func (h *ScanHistory) save() {
	data, err := json.MarshalIndent(h.data, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(h.filePath), 0755); err != nil {
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
