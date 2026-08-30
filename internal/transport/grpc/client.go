package grpc

import (
	"context"
	"fmt"

	"github.com/elip/WeaveLens/internal/application/service"
)

type GRPCClient struct {
	discoveryClient *DiscoveryClient
	graphClient     *GraphClient
}

type DiscoveryClient struct {
	service.DiscoveryService
}

func NewDiscoveryClient(discovery service.DiscoveryService) *DiscoveryClient {
	return &DiscoveryClient{DiscoveryService: discovery}
}

func (c *DiscoveryClient) StartScan(ctx context.Context, req *StartScanRequest) (*StartScanResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	scanID, err := c.DiscoveryService.StartScan(propagateContext(ctx), req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to start scan: %w", err)
	}

	return &StartScanResponse{
		ScanID:  scanID,
		Status:  "RUNNING",
		Message: "scan started",
	}, nil
}

func (c *DiscoveryClient) GetScanStatus(ctx context.Context, req *GetScanStatusRequest) (*GetScanStatusResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	status, count, err := c.DiscoveryService.GetScanStatus(propagateContext(ctx), req.ScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan status: %w", err)
	}

	return &GetScanStatusResponse{
		ScanID:       req.ScanID,
		Status:       status,
		ResourceCount: int32(count),
		Message:      "scan status retrieved",
	}, nil
}

func (c *DiscoveryClient) CancelScan(ctx context.Context, req *CancelScanRequest) (*CancelScanResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := c.DiscoveryService.CancelScan(propagateContext(ctx), req.ScanID); err != nil {
		return nil, fmt.Errorf("failed to cancel scan: %w", err)
	}

	return &CancelScanResponse{
		Cancelled: true,
		Message:   "scan cancelled",
	}, nil
}

func (c *DiscoveryClient) ListResources(ctx context.Context, req *ListResourcesRequest) (*ListResourcesResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resources, err := c.DiscoveryService.ListResources(propagateContext(ctx), req.ScanID, req.Category, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
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

type GraphClient struct {
	service.GraphService
}

func NewGraphClient(graph service.GraphService) *GraphClient {
	return &GraphClient{GraphService: graph}
}

func (c *GraphClient) GetGraph(ctx context.Context, req *GetGraphRequest) (*GetGraphResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	nodes, edges, err := c.GraphService.GetGraph(propagateContext(ctx), req.ScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
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

func (c *GraphClient) GetResource(ctx context.Context, req *GetResourceRequest) (*GetResourceResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resource, err := c.GraphService.GetResource(propagateContext(ctx), req.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	return &GetResourceResponse{
		Resource: toProtoResource(resource),
	}, nil
}

func (c *GraphClient) GetNeighbors(ctx context.Context, req *GetNeighborsRequest) (*GetNeighborsResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	neighbors, err := c.GraphService.GetNeighbors(propagateContext(ctx), req.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get neighbors: %w", err)
	}

	var protoNeighbors []Resource
	for _, n := range neighbors {
		protoNeighbors = append(protoNeighbors, toProtoResource(n))
	}

	return &GetNeighborsResponse{
		Neighbors: protoNeighbors,
	}, nil
}

func (c *GraphClient) GetRelationships(ctx context.Context, req *GetRelationshipsRequest) (*GetRelationshipsResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	relationships, err := c.GraphService.GetRelationships(propagateContext(ctx), req.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships: %w", err)
	}

	var protoRelationships []Relationship
	for _, rel := range relationships {
		protoRelationships = append(protoRelationships, toProtoRelationship(rel))
	}

	return &GetRelationshipsResponse{
		Relationships: protoRelationships,
	}, nil
}

func NewGRPCClient(discovery service.DiscoveryService, graph service.GraphService) *GRPCClient {
	return &GRPCClient{
		discoveryClient: NewDiscoveryClient(discovery),
		graphClient:     NewGraphClient(graph),
	}
}
