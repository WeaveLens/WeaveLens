package service

import (
	"context"
	"log/slog"
	"strings"
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
	Regions   []string
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

func (s *discoveryService) StartScan(ctx context.Context, regions []string) (string, error) {
	scanID := generateScanID()

	displayRegion := "all"
	if len(regions) == 1 {
		displayRegion = regions[0]
	} else if len(regions) > 1 {
		displayRegion = strings.Join(regions, ",")
	}

	s.mu.Lock()
	s.scans[scanID] = &ScanRecord{
		ID:        scanID,
		Status:    "RUNNING",
		Region:    displayRegion,
		Regions:   regions,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.mu.Unlock()

	if s.graphService != nil {
		s.graphService.SetScanRegions(scanID, regions)
	}

	if s.history != nil {
		s.history.AddScan(scanID, displayRegion, regions)
	}

	evt := &event.ScanStartedEvent{
		ScanID:  scanID,
		Region:  displayRegion,
		Regions: regions,
	}

	if err := s.eventBus.PublishScanStarted(ctx, evt); err != nil {
		s.logger.Error("failed to publish scan started event", "error", err, "scanID", scanID)
		return "", err
	}

	s.logger.Info("scan started", "scanID", scanID, "regions", regions)
	return scanID, nil
}

func (s *discoveryService) GetScanStatus(ctx context.Context, scanID string) (string, int, error) {
	s.mu.RLock()
	record, exists := s.scans[scanID]
	s.mu.RUnlock()

	if exists {
		return record.Status, 0, nil
	}

	if s.history != nil {
		scans := s.history.GetScans()
		for _, scan := range scans {
			if scan.ID == scanID {
				return scan.Status, scan.NodeCount, nil
			}
		}
	}

	return "NOT_FOUND", 0, nil
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

func (s *discoveryService) DeleteScan(ctx context.Context, scanID string) (bool, error) {
	if s.history == nil {
		return false, nil
	}

	scan, found := s.history.FindScan(scanID)
	if !found {
		return false, nil
	}

	if scan.Pinned {
		s.logger.Warn("attempted to delete pinned scan", "scanID", scanID)
		return false, nil
	}

	if scan.Status == "RUNNING" {
		s.mu.RLock()
		_, inMemory := s.scans[scanID]
		s.mu.RUnlock()
		if inMemory {
			s.logger.Warn("attempted to delete running scan", "scanID", scanID)
			return false, nil
		}
	}

	s.mu.Lock()
	delete(s.scans, scanID)
	s.mu.Unlock()

	if s.graphService != nil {
		s.graphService.SetScanRegions(scanID, nil)
	}

	removed := s.history.RemoveScan(scanID)
	if !removed {
		return false, nil
	}

	s.logger.Info("scan deleted", "scanID", scanID)
	return true, nil
}

func (s *discoveryService) SetScanPinned(ctx context.Context, scanID string, pinned bool) (bool, error) {
	if s.history == nil {
		return false, nil
	}
	return s.history.SetScanPinned(scanID, pinned), nil
}

func (s *discoveryService) SetScanLocked(ctx context.Context, scanID string, locked bool) (bool, error) {
	if s.history == nil {
		return false, nil
	}
	return s.history.SetScanLocked(scanID, locked), nil
}

func (s *discoveryService) ClearUnpinned(ctx context.Context) (int, error) {
	if s.history == nil {
		return 0, nil
	}

	scans := s.history.GetScans()
	removed := 0
	s.mu.Lock()
	for _, scan := range scans {
		if scan.Pinned {
			continue
		}
		delete(s.scans, scan.ID)
		if s.graphService != nil {
			s.graphService.SetScanRegions(scan.ID, nil)
		}
	}
	s.mu.Unlock()

	removed = s.history.RemoveUnpinned()
	s.logger.Info("cleared unpinned scans", "removed", removed)
	return removed, nil
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
