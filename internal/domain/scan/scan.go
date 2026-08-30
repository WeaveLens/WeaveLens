package scan

import (
	"time"

	"github.com/elip/WeaveLens/internal/domain/graph"
)

type ScanID string

type Scan struct {
	id        ScanID
	graph     *graph.Graph
	createdAt time.Time
	metadata  map[string]string
}

type ScanOption func(*Scan)

func WithScanMetadata(metadata map[string]string) ScanOption {
	return func(s *Scan) {
		s.metadata = metadata
	}
}

func WithScanTime(t time.Time) ScanOption {
	return func(s *Scan) {
		s.createdAt = t
	}
}

func NewScan(id ScanID, g *graph.Graph, opts ...ScanOption) *Scan {
	s := &Scan{
		id:        id,
		graph:     g,
		createdAt: time.Now(),
		metadata:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Scan) ID() ScanID {
	return s.id
}

func (s *Scan) Graph() *graph.Graph {
	return s.graph
}

func (s *Scan) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Scan) Metadata() map[string]string {
	return s.metadata
}
