package discovery

import (
	"context"
	"errors"
	"testing"

	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type mockScanner struct {
	resources []*resource.Resource
	err       error
}

func (m *mockScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	return m.resources, m.err
}

type mockRelationshipBuilder struct {
	relationships []*relationship.Relationship
	err           error
}

func (m *mockRelationshipBuilder) Build(resources []*resource.Resource) ([]*relationship.Relationship, error) {
	return m.relationships, m.err
}

func TestService_Discover(t *testing.T) {
	vpc, _ := resource.NewResource("vpc-123", "VPC", resource.CategoryNetwork, "test-vpc")

	scanner := &mockScanner{
		resources: []*resource.Resource{vpc},
	}
	rb := &mockRelationshipBuilder{
		relationships: []*relationship.Relationship{},
	}

	service := NewService([]Scanner{scanner}, rb)
	result, err := service.Discover(context.Background(), DiscoveryRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("Discover() unexpected error: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(result.Resources))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestService_DiscoverWithScannerError(t *testing.T) {
	scanner := &mockScanner{
		err: errors.New("scanner failed"),
	}
	rb := &mockRelationshipBuilder{}

	service := NewService([]Scanner{scanner}, rb)
	result, err := service.Discover(context.Background(), DiscoveryRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("Discover() unexpected error: %v", err)
	}

	if len(result.Resources) != 0 {
		t.Errorf("Expected 0 resources, got %d", len(result.Resources))
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}

func TestService_DiscoverWithContextCancellation(t *testing.T) {
	scanner := &mockScanner{
		resources: []*resource.Resource{},
	}
	rb := &mockRelationshipBuilder{}

	service := NewService([]Scanner{scanner}, rb)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.Discover(ctx, DiscoveryRequest{Region: "us-east-1"})
	if err == nil {
		t.Fatal("Discover() expected error for cancelled context")
	}
}

func TestService_DiscoverPartialFailure(t *testing.T) {
	vpc, _ := resource.NewResource("vpc-123", "VPC", resource.CategoryNetwork, "test-vpc")

	scanner1 := &mockScanner{
		resources: []*resource.Resource{vpc},
	}
	scanner2 := &mockScanner{
		err: errors.New("scanner 2 failed"),
	}
	rb := &mockRelationshipBuilder{
		relationships: []*relationship.Relationship{},
	}

	service := NewService([]Scanner{scanner1, scanner2}, rb)
	result, err := service.Discover(context.Background(), DiscoveryRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("Discover() unexpected error: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(result.Resources))
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}
