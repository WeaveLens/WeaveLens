package service

import "context"

type DiscoveryService interface {
	StartScan(ctx context.Context, region string) (string, error)
	GetScanStatus(ctx context.Context, scanID string) (string, int, error)
	CancelScan(ctx context.Context, scanID string) error
	ListResources(ctx context.Context, scanID, category, resourceType string) ([]Resource, error)
}

type GraphService interface {
	GetGraph(ctx context.Context, scanID string) ([]Resource, []Relationship, error)
	GetResource(ctx context.Context, resourceID string) (Resource, error)
	GetNeighbors(ctx context.Context, resourceID string) ([]Resource, error)
	GetRelationships(ctx context.Context, resourceID string) ([]Relationship, error)
}

type Resource struct {
	ID       string
	Name     string
	Type     string
	Category string
	ARN      string
	Region   string
	Metadata map[string]string
}

type Relationship struct {
	ID       string
	SourceID string
	TargetID string
	Type     string
	Metadata map[string]string
}
