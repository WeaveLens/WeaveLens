package grpc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elip/WeaveLens/internal/application/service"
)

type ScanRecord struct {
	ID        string
	Status    string
	Region    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DiscoveryServer struct {
	service.DiscoveryService
	mu    sync.RWMutex
	scans map[string]*ScanRecord
}

func NewDiscoveryServer(discovery service.DiscoveryService) *DiscoveryServer {
	return &DiscoveryServer{
		DiscoveryService: discovery,
		scans:            make(map[string]*ScanRecord),
	}
}

func (s *DiscoveryServer) StartScan(ctx context.Context, req *StartScanRequest) (*StartScanResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	scanID, err := s.DiscoveryService.StartScan(ctx, req.Region)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	s.mu.Lock()
	s.scans[scanID] = &ScanRecord{
		ID:        scanID,
		Status:    "RUNNING",
		Region:    req.Region,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.mu.Unlock()

	return &StartScanResponse{
		ScanID:  scanID,
		Status:  "RUNNING",
		Message: "scan started",
	}, nil
}

func (s *DiscoveryServer) GetScanStatus(ctx context.Context, req *GetScanStatusRequest) (*GetScanStatusResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	status, count, err := s.DiscoveryService.GetScanStatus(ctx, req.ScanID)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	s.mu.Lock()
	if record, exists := s.scans[req.ScanID]; exists {
		record.Status = status
		record.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	return &GetScanStatusResponse{
		ScanID:        req.ScanID,
		Status:        status,
		ResourceCount: int32(count),
		Message:       "scan status retrieved",
	}, nil
}

func (s *DiscoveryServer) CancelScan(ctx context.Context, req *CancelScanRequest) (*CancelScanResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	if err := s.DiscoveryService.CancelScan(ctx, req.ScanID); err != nil {
		return nil, mapGRPCError(err)
	}

	s.mu.Lock()
	if record, exists := s.scans[req.ScanID]; exists {
		record.Status = "CANCELLED"
		record.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	return &CancelScanResponse{
		Cancelled: true,
		Message:   "scan cancelled",
	}, nil
}

func (s *DiscoveryServer) ListResources(ctx context.Context, req *ListResourcesRequest) (*ListResourcesResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	resources, err := s.DiscoveryService.ListResources(ctx, req.ScanID, req.Category, req.Type)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	var protoResources []Resource
	for _, r := range resources {
		protoResources = append(protoResources, toProtoResource(r))
	}

	return &ListResourcesResponse{
		Resources: protoResources,
		Total:     int32(len(protoResources)),
	}, nil
}

type GraphServer struct {
	service.GraphService
}

func NewGraphServer(graph service.GraphService) *GraphServer {
	return &GraphServer{GraphService: graph}
}

func (s *GraphServer) GetGraph(ctx context.Context, req *GetGraphRequest) (*GetGraphResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	nodes, edges, err := s.GraphService.GetGraph(ctx, req.ScanID)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	var protoNodes []Resource
	for _, n := range nodes {
		protoNodes = append(protoNodes, toProtoResource(n))
	}

	var protoEdges []Relationship
	for _, e := range edges {
		protoEdges = append(protoEdges, toProtoRelationship(e))
	}

	return &GetGraphResponse{
		Nodes:     protoNodes,
		Edges:     protoEdges,
		NodeCount: int32(len(protoNodes)),
		EdgeCount: int32(len(protoEdges)),
	}, nil
}

func (s *GraphServer) GetResource(ctx context.Context, req *GetResourceRequest) (*GetResourceResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	resource, err := s.GraphService.GetResource(ctx, req.ResourceID)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	return &GetResourceResponse{
		Resource: toProtoResource(resource),
	}, nil
}

func (s *GraphServer) GetNeighbors(ctx context.Context, req *GetNeighborsRequest) (*GetNeighborsResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	neighbors, err := s.GraphService.GetNeighbors(ctx, req.ResourceID)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	var protoNeighbors []Resource
	for _, n := range neighbors {
		protoNeighbors = append(protoNeighbors, toProtoResource(n))
	}

	return &GetNeighborsResponse{
		Neighbors: protoNeighbors,
	}, nil
}

func (s *GraphServer) GetRelationships(ctx context.Context, req *GetRelationshipsRequest) (*GetRelationshipsResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	relationships, err := s.GraphService.GetRelationships(ctx, req.ResourceID)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	var protoRelationships []Relationship
	for _, rel := range relationships {
		protoRelationships = append(protoRelationships, toProtoRelationship(rel))
	}

	return &GetRelationshipsResponse{
		Relationships: protoRelationships,
	}, nil
}

func validateRequest(req interface{}) error {
	switch r := req.(type) {
	case *StartScanRequest:
		// Empty region means "all regions" and is valid for cross-region scans.
		return nil
	case *GetScanStatusRequest:
		if strings.TrimSpace(r.ScanID) == "" {
			return fmt.Errorf("scan_id must not be empty")
		}
	case *CancelScanRequest:
		if strings.TrimSpace(r.ScanID) == "" {
			return fmt.Errorf("scan_id must not be empty")
		}
	case *ListResourcesRequest:
		if strings.TrimSpace(r.ScanID) == "" {
			return fmt.Errorf("scan_id must not be empty")
		}
	case *GetGraphRequest:
		if strings.TrimSpace(r.ScanID) == "" {
			return fmt.Errorf("scan_id must not be empty")
		}
	case *GetResourceRequest:
		if strings.TrimSpace(r.ResourceID) == "" {
			return fmt.Errorf("resource_id must not be empty")
		}
	case *GetNeighborsRequest:
		if strings.TrimSpace(r.ResourceID) == "" {
			return fmt.Errorf("resource_id must not be empty")
		}
	case *GetRelationshipsRequest:
		if strings.TrimSpace(r.ResourceID) == "" {
			return fmt.Errorf("resource_id must not be empty")
		}
	}
	return nil
}

func mapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("grpc_error: %w", err)
}

func propagateContext(ctx context.Context) context.Context {
	return ctx
}

func toProtoResource(r service.Resource) Resource {
	return Resource{
		ID:       r.ID,
		Name:     r.Name,
		Type:     r.Type,
		Category: r.Category,
		ARN:      r.ARN,
		Region:   r.Region,
		Metadata: r.Metadata,
	}
}

func toProtoRelationship(r service.Relationship) Relationship {
	return Relationship{
		ID:       r.ID,
		SourceID: r.SourceID,
		TargetID: r.TargetID,
		Type:     r.Type,
		Metadata: r.Metadata,
	}
}
