package grpc

import "time"

type Resource struct {
	ID         string
	ARN        string
	AccountID  string
	Region     string
	Type       string
	Category   string
	Name       string
	Metadata   map[string]string
}

type Relationship struct {
	ID       string
	SourceID string
	TargetID string
	Type     string
	Metadata map[string]string
}

type Scan struct {
	ID        string
	Status    string
	Resources []Resource
	CreatedAt time.Time
	UpdatedAt time.Time
}

type StartScanRequest struct {
	Region  string
	Regions []string
}

type StartScanResponse struct {
	ScanID  string
	Status  string
	Message string
}

type GetScanStatusRequest struct {
	ScanID string
}

type GetScanStatusResponse struct {
	ScanID       string
	Status       string
	ResourceCount int32
	Message      string
}

type CancelScanRequest struct {
	ScanID string
}

type CancelScanResponse struct {
	Cancelled bool
	Message   string
}

type ListResourcesRequest struct {
	ScanID  string
	Category string
	Type    string
}

type ListResourcesResponse struct {
	Resources []Resource
	Total     int32
}

type GetGraphRequest struct {
	ScanID string
}

type GetGraphResponse struct {
	Nodes     []Resource
	Edges     []Relationship
	NodeCount int32
	EdgeCount int32
}

type GetResourceRequest struct {
	ResourceID string
}

type GetResourceResponse struct {
	Resource Resource
}

type GetNeighborsRequest struct {
	ResourceID string
}

type GetNeighborsResponse struct {
	Neighbors []Resource
}

type GetRelationshipsRequest struct {
	ResourceID string
}

type GetRelationshipsResponse struct {
	Relationships []Relationship
}
