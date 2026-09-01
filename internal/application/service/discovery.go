package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/elip/WeaveLens/internal/domain/event"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/discovery"
	"github.com/elip/WeaveLens/internal/infrastructure/nats"
)

type discoveryService struct {
	eventBus  *nats.EventBus
	logger    *slog.Logger
	mu        sync.RWMutex
	scans     map[string]*ScanRecord
	discovery discovery.ResourceDiscovery
	graphService GraphService
	history *ScanHistory
}

func (s *discoveryService) SetGraphService(gs GraphService) {
	s.graphService = gs
}

func (s *discoveryService) SetHistory(h *ScanHistory) {
	s.history = h
}

func (s *discoveryService) GetScans() []ScanHistoryEntry {
	if s.history == nil {
		return []ScanHistoryEntry{}
	}
	return s.history.GetScans()
}

type ScanRecord struct {
	ID        string
	Status    string
	Region    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDiscoveryService(eventBus *nats.EventBus, logger *slog.Logger, discovery discovery.ResourceDiscovery) DiscoveryService {
	return &discoveryService{
		eventBus:  eventBus,
		logger:    logger,
		scans:     make(map[string]*ScanRecord),
		discovery: discovery,
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

	if s.graphService != nil {
		s.graphService.SetScanRegion(scanID, region)
	}

	if s.history != nil {
		s.history.AddScan(scanID, region)
	}

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

func (s *discoveryService) CompleteScan(ctx context.Context, scanID string, nodeCount, edgeCount int) error {
	s.mu.Lock()
	record, exists := s.scans[scanID]
	if exists {
		record.Status = "COMPLETED"
		record.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	if !exists {
		return nil
	}

	if s.history != nil {
		s.history.UpdateScan(scanID, "COMPLETED", nodeCount, edgeCount)
	}

	evt := &event.ScanCompletedEvent{
		ScanID:        scanID,
		ResourceCount: nodeCount,
	}

	if err := s.eventBus.PublishScanCompleted(ctx, evt); err != nil {
		s.logger.Error("failed to publish scan completed event", "error", err, "scanID", scanID)
	}

	s.logger.Info("scan completed", "scanID", scanID, "resources", nodeCount, "relationships", edgeCount)
	return nil
}

func (s *discoveryService) ListResources(ctx context.Context, scanID, category, resourceType string) ([]Resource, error) {
	if s.discovery == nil {
		return []Resource{}, nil
	}

	result, err := s.discovery.Discover(ctx, discovery.DiscoveryRequest{Region: ""})
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for _, res := range result.Resources {
		resources = append(resources, Resource{
			ID:       string(res.ID()),
			Name:     res.Name(),
			Type:     string(res.Type()),
			Category: string(res.Category()),
			ARN:      res.ARN(),
			Region:   res.Region(),
			Metadata: res.Metadata(),
			Tags:     res.Tags(),
		})
	}
	return resources, nil
}

func generateScanID() string {
	return "scan-" + time.Now().Format("20060102150405")
}
