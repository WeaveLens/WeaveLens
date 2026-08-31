package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/elip/WeaveLens/internal/domain/event"
	"github.com/elip/WeaveLens/internal/infrastructure/nats"
)

type discoveryService struct {
	eventBus *nats.EventBus
	logger   *slog.Logger
	mu       sync.RWMutex
	scans    map[string]*ScanRecord
}

type ScanRecord struct {
	ID        string
	Status    string
	Region    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDiscoveryService(eventBus *nats.EventBus, logger *slog.Logger) DiscoveryService {
	return &discoveryService{
		eventBus: eventBus,
		logger:   logger,
		scans:    make(map[string]*ScanRecord),
	}
}

func (s *discoveryService) StartScan(ctx context.Context, region string) (string, error) {
	scanID := generateScanID()

	s.mu.Lock()
	s.scans[scanID] = &ScanRecord{
		ID:        scanID,
		Status:    "RUNNING",
		Region:    region,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.mu.Unlock()

	evt := &event.ScanStartedEvent{
		ScanID: scanID,
		Region: region,
	}

	if err := s.eventBus.PublishScanStarted(ctx, evt); err != nil {
		s.logger.Error("failed to publish scan started event", "error", err, "scanID", scanID)
		return "", err
	}

	s.logger.Info("scan started", "scanID", scanID, "region", region)
	return scanID, nil
}

func (s *discoveryService) GetScanStatus(ctx context.Context, scanID string) (string, int, error) {
	s.mu.RLock()
	record, exists := s.scans[scanID]
	s.mu.RUnlock()

	if !exists {
		return "NOT_FOUND", 0, nil
	}

	return record.Status, 0, nil
}

func (s *discoveryService) CancelScan(ctx context.Context, scanID string) error {
	s.mu.Lock()
	record, exists := s.scans[scanID]
	if exists {
		record.Status = "CANCELLED"
		record.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	if !exists {
		return nil
	}

	s.logger.Info("scan cancelled", "scanID", scanID)
	return nil
}

func (s *discoveryService) ListResources(ctx context.Context, scanID, category, resourceType string) ([]Resource, error) {
	return []Resource{
		{ID: "res-1", Name: "test", Type: "EC2", Category: "compute"},
	}, nil
}

type graphService struct {
	eventBus *nats.EventBus
	logger   *slog.Logger
}

func NewGraphService(eventBus *nats.EventBus, logger *slog.Logger) GraphService {
	return &graphService{
		eventBus: eventBus,
		logger:   logger,
	}
}

func (s *graphService) GetGraph(ctx context.Context, scanID string) ([]Resource, []Relationship, error) {
	return []Resource{
		{ID: "res-1", Name: "test", Type: "EC2", Category: "compute"},
	}, []Relationship{
		{ID: "rel-1", SourceID: "vpc-1", TargetID: "subnet-1", Type: "contains"},
	}, nil
}

func (s *graphService) GetResource(ctx context.Context, resourceID string) (Resource, error) {
	return Resource{ID: resourceID, Name: "test", Type: "EC2", Category: "compute"}, nil
}

func (s *graphService) GetNeighbors(ctx context.Context, resourceID string) ([]Resource, error) {
	return []Resource{
		{ID: "res-2", Name: "neighbor", Type: "Subnet", Category: "network"},
	}, nil
}

func (s *graphService) GetRelationships(ctx context.Context, resourceID string) ([]Relationship, error) {
	return []Relationship{
		{ID: "rel-1", SourceID: resourceID, TargetID: "res-2", Type: "contains"},
	}, nil
}

func generateScanID() string {
	return "scan-" + time.Now().Format("20060102150405")
}
