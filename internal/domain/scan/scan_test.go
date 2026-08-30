package scan_test

import (
	"testing"
	"time"

	"github.com/elip/WeaveLens/internal/domain/graph"
	"github.com/elip/WeaveLens/internal/domain/scan"
)

func TestNewScan(t *testing.T) {
	g := graph.NewGraph()
	s := scan.NewScan("scan-1", g)

	if s.ID() != "scan-1" {
		t.Errorf("ID() = %v, want scan-1", s.ID())
	}

	if s.Graph() != g {
		t.Errorf("Graph() = %v, want %v", s.Graph(), g)
	}

	if s.CreatedAt().IsZero() {
		t.Errorf("CreatedAt() is zero")
	}
}

func TestScanWithOptions(t *testing.T) {
	g := graph.NewGraph()
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := scan.NewScan(
		"scan-1",
		g,
		scan.WithScanTime(timestamp),
		scan.WithScanMetadata(map[string]string{"env": "prod"}),
	)

	if !s.CreatedAt().Equal(timestamp) {
		t.Errorf("CreatedAt() = %v, want %v", s.CreatedAt(), timestamp)
	}

	meta := s.Metadata()
	if meta["env"] != "prod" {
		t.Errorf("Metadata['env'] = %v, want prod", meta["env"])
	}
}

func TestScanIdentity(t *testing.T) {
	g := graph.NewGraph()
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s1 := scan.NewScan("scan-1", g, scan.WithScanTime(timestamp))
	s2 := scan.NewScan("scan-1", g, scan.WithScanTime(timestamp))

	if s1.ID() != s2.ID() {
		t.Errorf("Scan IDs should be equal for same ID input")
	}

	if !s1.CreatedAt().Equal(s2.CreatedAt()) {
		t.Errorf("CreatedAt() should be independent for different scans")
	}
}
