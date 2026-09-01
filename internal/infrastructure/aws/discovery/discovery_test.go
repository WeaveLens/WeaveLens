package discovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/discovery/resilience"
)

type mockScanner struct {
	name      string
	resources []*resource.Resource
	err       error
}

func (m *mockScanner) Name() string {
	return m.name
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
		name:      "test",
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
		name: "test",
		err:  errors.New("scanner failed"),
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
		name:      "test",
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
		name:      "scanner1",
		resources: []*resource.Resource{vpc},
	}
	scanner2 := &mockScanner{
		name: "scanner2",
		err:  errors.New("scanner 2 failed"),
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

func TestService_DiscoverConcurrent(t *testing.T) {
	vpc, _ := resource.NewResource("vpc-123", "VPC", resource.CategoryNetwork, "test-vpc")
	ec2, _ := resource.NewResource("i-456", "EC2", resource.CategoryCompute, "test-ec2")

	scanner1 := &mockScanner{
		name:      "network",
		resources: []*resource.Resource{vpc},
	}
	scanner2 := &mockScanner{
		name:      "compute",
		resources: []*resource.Resource{ec2},
	}
	rb := &mockRelationshipBuilder{
		relationships: []*relationship.Relationship{},
	}

	config := ServiceConfig{
		Workers:     4,
		RetryConfig: resilience.RetryConfig{MaxAttempts: 0},
	}
	service := NewServiceWithConfig([]Scanner{scanner1, scanner2}, rb, config)
	result, err := service.Discover(context.Background(), DiscoveryRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("Discover() unexpected error: %v", err)
	}

	if len(result.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(result.Resources))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestService_DiscoverWithRetry(t *testing.T) {
	vpc, _ := resource.NewResource("vpc-123", "VPC", resource.CategoryNetwork, "test-vpc")

	scanner := &mockScanner{
		name:      "test",
		resources: []*resource.Resource{vpc},
	}
	rb := &mockRelationshipBuilder{
		relationships: []*relationship.Relationship{},
	}

	config := ServiceConfig{
		Workers: 2,
		RetryConfig: resilience.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			JitterFactor: 0,
		},
	}
	service := NewServiceWithConfig([]Scanner{scanner}, rb, config)
	result, err := service.Discover(context.Background(), DiscoveryRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("Discover() unexpected error: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(result.Resources))
	}
}

func TestService_WorkerLimit(t *testing.T) {
	var scanners []Scanner
	for i := 0; i < 10; i++ {
		scanners = append(scanners, &mockScanner{
			name:      "scanner",
			resources: []*resource.Resource{},
		})
	}
	rb := &mockRelationshipBuilder{}

	config := ServiceConfig{
		Workers:     3,
		RetryConfig: resilience.RetryConfig{MaxAttempts: 0},
	}
	service := NewServiceWithConfig(scanners, rb, config)

	result, err := service.Discover(context.Background(), DiscoveryRequest{Region: "us-east-1"})
	if err != nil {
		t.Fatalf("Discover() unexpected error: %v", err)
	}

	if len(result.Resources) != 0 {
		t.Errorf("Expected 0 resources, got %d", len(result.Resources))
	}
}
