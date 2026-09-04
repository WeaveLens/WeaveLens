package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const positionsFile = "positions.json"

type PositionsStore struct {
	mu       sync.RWMutex
	filePath string
	data     map[string]PositionData
}

type PositionData struct {
	Positions map[string]struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"positions"`
	Viewport *struct {
		Zoom float64 `json:"zoom"`
		Pan  struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"pan"`
	} `json:"viewport,omitempty"`
}

func NewPositionsStore() *PositionsStore {
	p := &PositionsStore{
		data:     make(map[string]PositionData),
		filePath: resolvePositionsPath(),
	}
	p.load()
	return p
}

func resolvePositionsPath() string {
	if path := os.Getenv("WEAVELENS_POSITIONS_FILE"); path != "" {
		return path
	}
	dir, err := os.Getwd()
	if err != nil {
		return filepath.Join(".", "data", positionsFile)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "data", positionsFile)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join(".", "data", positionsFile)
		}
		dir = parent
	}
}

func (p *PositionsStore) Get(scanID string) (PositionData, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.data[scanID]
	return val, ok
}

func (p *PositionsStore) Save(scanID string, data PositionData) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[scanID] = data
	p.save()
}

func (p *PositionsStore) Delete(scanID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.data, scanID)
	p.save()
}

func (p *PositionsStore) load() {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		p.data = make(map[string]PositionData)
		return
	}
	if err := json.Unmarshal(data, &p.data); err != nil {
		p.data = make(map[string]PositionData)
	}
	if p.data == nil {
		p.data = make(map[string]PositionData)
	}
}

func (p *PositionsStore) save() {
	data, err := json.MarshalIndent(p.data, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p.filePath), 0755); err != nil {
		return
	}
	_ = os.WriteFile(p.filePath, data, 0644)
}
