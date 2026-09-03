package service

import "context"

type DiscoveryService interface {
	StartScan(ctx context.Context, regions []string) (string, error)
	GetScanStatus(ctx context.Context, scanID string) (string, int, error)
	CancelScan(ctx context.Context, scanID string) error
	ListResources(ctx context.Context, scanID, category, resourceType string) ([]Resource, error)
	CompleteScan(ctx context.Context, scanID string, nodeCount, edgeCount int) error
	SetGraphService(gs GraphService)
	SetHistory(h *ScanHistory)
	GetScans() []ScanHistoryEntry
}

type GraphService interface {
	GetGraph(ctx context.Context, scanID string) ([]Resource, []Relationship, error)
	GetResource(ctx context.Context, resourceID string) (Resource, error)
	GetNeighbors(ctx context.Context, resourceID string) ([]Resource, error)
	GetRelationships(ctx context.Context, resourceID string) ([]Relationship, error)
	SetScanRegions(scanID string, regions []string)
	SetHistory(h *ScanHistory)
}

type Resource struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Category string            `json:"category"`
	ARN      string            `json:"arn"`
	Region   string            `json:"region"`
	Metadata map[string]string `json:"metadata"`
	Tags     map[string]string `json:"tags"`
}

type Relationship struct {
	ID       string            `json:"id"`
	SourceID string            `json:"sourceId"`
	TargetID string            `json:"targetId"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
}
