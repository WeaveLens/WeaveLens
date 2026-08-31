package discovery

import (
	"testing"

	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

func TestRelationshipBuilder_Build(t *testing.T) {
	vpc, _ := resource.NewResource("vpc-123", "VPC", resource.CategoryNetwork, "test-vpc")
	subnet, _ := resource.NewResource("subnet-456", "Subnet", resource.CategoryNetwork, "test-subnet",
		resource.WithMetadata(map[string]string{"vpc_id": "vpc-123"}))
	ec2, _ := resource.NewResource("i-789", "EC2", resource.CategoryCompute, "test-ec2",
		resource.WithMetadata(map[string]string{"subnet_id": "subnet-456"}))

	resources := []*resource.Resource{vpc, subnet, ec2}
	builder := NewRelationshipBuilder()
	relationships, err := builder.Build(resources)
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}

	if len(relationships) != 2 {
		t.Errorf("Expected 2 relationships, got %d", len(relationships))
	}

	foundVpcSubnet := false
	foundSubnetEC2 := false
	for _, rel := range relationships {
		if rel.SourceID() == "vpc-123" && rel.TargetID() == "subnet-456" && rel.Type() == relationship.RelationshipContains {
			foundVpcSubnet = true
		}
		if rel.SourceID() == "subnet-456" && rel.TargetID() == "i-789" && rel.Type() == relationship.RelationshipContains {
			foundSubnetEC2 = true
		}
	}

	if !foundVpcSubnet {
		t.Error("Expected VPC->Subnet relationship")
	}
	if !foundSubnetEC2 {
		t.Error("Expected Subnet->EC2 relationship")
	}
}

func TestRelationshipBuilder_EmptyResources(t *testing.T) {
	builder := NewRelationshipBuilder()
	relationships, err := builder.Build([]*resource.Resource{})
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	if len(relationships) != 0 {
		t.Errorf("Expected 0 relationships, got %d", len(relationships))
	}
}

func TestRelationshipBuilder_OrphanedResources(t *testing.T) {
	subnet, _ := resource.NewResource("subnet-456", "Subnet", resource.CategoryNetwork, "test-subnet",
		resource.WithMetadata(map[string]string{"vpc_id": "nonexistent-vpc"}))

	resources := []*resource.Resource{subnet}
	builder := NewRelationshipBuilder()
	relationships, err := builder.Build(resources)
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	if len(relationships) != 0 {
		t.Errorf("Expected 0 relationships for orphaned subnet, got %d", len(relationships))
	}
}
