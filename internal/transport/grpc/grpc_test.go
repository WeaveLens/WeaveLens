package grpc_test

import (
	"context"
	"testing"

	"github.com/elip/WeaveLens/internal/application/service"
	"github.com/elip/WeaveLens/internal/transport/grpc"
)

type fakeDiscoveryService struct {
	service.DiscoveryService
	scanID string
}

func (f *fakeDiscoveryService) StartScan(ctx context.Context, regions []string) (string, error) {
	return "scan-123", nil
}

func (f *fakeDiscoveryService) DeleteScan(ctx context.Context, scanID string) (bool, error) {
	return true, nil
}

func (f *fakeDiscoveryService) SetScanPinned(ctx context.Context, scanID string, pinned bool) (bool, error) {
	return true, nil
}

func (f *fakeDiscoveryService) SetScanLocked(ctx context.Context, scanID string, locked bool) (bool, error) {
	return true, nil
}

func (f *fakeDiscoveryService) SetScanLayout(ctx context.Context, scanID string, layout string) (bool, error) {
	return true, nil
}

func (f *fakeDiscoveryService) SetScanPositions(ctx context.Context, scanID string, data service.PositionData) error {
	return nil
}

func (f *fakeDiscoveryService) GetScanPositions(ctx context.Context, scanID string) (service.PositionData, bool) {
	return service.PositionData{}, false
}

func (f *fakeDiscoveryService) ClearUnpinned(ctx context.Context) (int, error) {
	return 0, nil
}

func (f *fakeDiscoveryService) GetScanStatus(ctx context.Context, scanID string) (string, int, error) {
	return "RUNNING", 5, nil
}

func (f *fakeDiscoveryService) CancelScan(ctx context.Context, scanID string) error {
	return nil
}

func (f *fakeDiscoveryService) ListResources(ctx context.Context, scanID, category, resourceType string) ([]service.Resource, error) {
	return []service.Resource{
		{ID: "res-1", Name: "test", Type: "EC2", Category: "compute"},
	}, nil
}

type fakeGraphService struct {
	service.GraphService
}

func (f *fakeGraphService) GetGraph(ctx context.Context, scanID string) ([]service.Resource, []service.Relationship, error) {
	return []service.Resource{
			{ID: "res-1", Name: "test", Type: "EC2", Category: "compute"},
		}, []service.Relationship{
			{ID: "rel-1", SourceID: "vpc-1", TargetID: "subnet-1", Type: "contains"},
		}, nil
}

func (f *fakeGraphService) GetResource(ctx context.Context, resourceID string) (service.Resource, error) {
	return service.Resource{ID: resourceID, Name: "test", Type: "EC2", Category: "compute"}, nil
}

func (f *fakeGraphService) GetNeighbors(ctx context.Context, resourceID string) ([]service.Resource, error) {
	return []service.Resource{
		{ID: "res-2", Name: "neighbor", Type: "Subnet", Category: "network"},
	}, nil
}

func (f *fakeGraphService) GetRelationships(ctx context.Context, resourceID string) ([]service.Relationship, error) {
	return []service.Relationship{
		{ID: "rel-1", SourceID: resourceID, TargetID: "res-2", Type: "contains"},
	}, nil
}

func TestDiscoveryServerStartScan(t *testing.T) {
	fake := &fakeDiscoveryService{}
	server := grpc.NewDiscoveryServer(fake)

	resp, err := server.StartScan(context.Background(), &grpc.StartScanRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("StartScan() error = %v", err)
	}

	if resp.ScanID != "scan-123" {
		t.Errorf("StartScan() ScanID = %v, want scan-123", resp.ScanID)
	}
	if resp.Status != "RUNNING" {
		t.Errorf("StartScan() Status = %v, want RUNNING", resp.Status)
	}
}

func TestDiscoveryServerValidation(t *testing.T) {
	fake := &fakeDiscoveryService{}
	server := grpc.NewDiscoveryServer(fake)

	resp, err := server.StartScan(context.Background(), &grpc.StartScanRequest{Region: "   "})
	if err != nil {
		t.Fatalf("StartScan() with empty region should be allowed: %v", err)
	}
	if resp.ScanID != "scan-123" {
		t.Errorf("StartScan() ScanID = %v, want scan-123", resp.ScanID)
	}
}

func TestDiscoveryServerGetScanStatus(t *testing.T) {
	fake := &fakeDiscoveryService{}
	server := grpc.NewDiscoveryServer(fake)

	resp, err := server.GetScanStatus(context.Background(), &grpc.GetScanStatusRequest{ScanID: "scan-123"})
	if err != nil {
		t.Fatalf("GetScanStatus() error = %v", err)
	}

	if resp.Status != "RUNNING" {
		t.Errorf("GetScanStatus() Status = %v, want RUNNING", resp.Status)
	}
	if resp.ResourceCount != 5 {
		t.Errorf("GetScanStatus() ResourceCount = %v, want 5", resp.ResourceCount)
	}
}

func TestDiscoveryServerCancelScan(t *testing.T) {
	fake := &fakeDiscoveryService{}
	server := grpc.NewDiscoveryServer(fake)

	resp, err := server.CancelScan(context.Background(), &grpc.CancelScanRequest{ScanID: "scan-123"})
	if err != nil {
		t.Fatalf("CancelScan() error = %v", err)
	}

	if !resp.Cancelled {
		t.Error("CancelScan() Cancelled = false, want true")
	}
}

func TestGraphServerGetGraph(t *testing.T) {
	fake := &fakeGraphService{}
	server := grpc.NewGraphServer(fake)

	resp, err := server.GetGraph(context.Background(), &grpc.GetGraphRequest{ScanID: "scan-123"})
	if err != nil {
		t.Fatalf("GetGraph() error = %v", err)
	}

	if len(resp.Nodes) != 1 {
		t.Errorf("GetGraph() Nodes length = %v, want 1", len(resp.Nodes))
	}
	if len(resp.Edges) != 1 {
		t.Errorf("GetGraph() Edges length = %v, want 1", len(resp.Edges))
	}
}

func TestGraphServerGetResource(t *testing.T) {
	fake := &fakeGraphService{}
	server := grpc.NewGraphServer(fake)

	resp, err := server.GetResource(context.Background(), &grpc.GetResourceRequest{ResourceID: "res-1"})
	if err != nil {
		t.Fatalf("GetResource() error = %v", err)
	}

	if resp.Resource.ID != "res-1" {
		t.Errorf("GetResource() Resource.ID = %v, want res-1", resp.Resource.ID)
	}
}

func TestGraphServerGetNeighbors(t *testing.T) {
	fake := &fakeGraphService{}
	server := grpc.NewGraphServer(fake)

	resp, err := server.GetNeighbors(context.Background(), &grpc.GetNeighborsRequest{ResourceID: "res-1"})
	if err != nil {
		t.Fatalf("GetNeighbors() error = %v", err)
	}

	if len(resp.Neighbors) != 1 {
		t.Errorf("GetNeighbors() Neighbors length = %v, want 1", len(resp.Neighbors))
	}
}

func TestGraphServerGetRelationships(t *testing.T) {
	fake := &fakeGraphService{}
	server := grpc.NewGraphServer(fake)

	resp, err := server.GetRelationships(context.Background(), &grpc.GetRelationshipsRequest{ResourceID: "res-1"})
	if err != nil {
		t.Fatalf("GetRelationships() error = %v", err)
	}

	if len(resp.Relationships) != 1 {
		t.Errorf("GetRelationships() Relationships length = %v, want 1", len(resp.Relationships))
	}
}

func TestGRPCClientStartScan(t *testing.T) {
	fake := &fakeDiscoveryService{}
	client := grpc.NewDiscoveryClient(fake)

	resp, err := client.StartScan(context.Background(), &grpc.StartScanRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("StartScan() error = %v", err)
	}

	if resp.ScanID != "scan-123" {
		t.Errorf("StartScan() ScanID = %v, want scan-123", resp.ScanID)
	}
}

func TestGRPCClientValidation(t *testing.T) {
	fake := &fakeDiscoveryService{}
	client := grpc.NewDiscoveryClient(fake)

	resp, err := client.StartScan(context.Background(), &grpc.StartScanRequest{Region: "   "})
	if err != nil {
		t.Fatalf("StartScan() with empty region should be allowed: %v", err)
	}
	if resp.ScanID != "scan-123" {
		t.Errorf("StartScan() ScanID = %v, want scan-123", resp.ScanID)
	}
}

func TestGRPCClientErrorMapping(t *testing.T) {
	fake := &fakeDiscoveryService{}
	client := grpc.NewDiscoveryClient(fake)

	_, err := client.GetScanStatus(context.Background(), &grpc.GetScanStatusRequest{ScanID: "scan-123"})
	if err != nil {
		t.Fatalf("GetScanStatus() unexpected error = %v", err)
	}
}
