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

func TestRelationshipBuilder_AttachmentMetadata(t *testing.T) {
	vpc, _ := resource.NewResource("vpc-123", "VPC", resource.CategoryNetwork, "vpc")
	subnet, _ := resource.NewResource("subnet-456", "Subnet", resource.CategoryNetwork, "subnet")
	securityGroup, _ := resource.NewResource("sg-789", "SecurityGroup", resource.CategorySecurity, "sg")
	routeTable, _ := resource.NewResource("rtb-012", "RouteTable", resource.CategoryNetwork, "route-table")
	alb, _ := resource.NewResource("alb-345", "ALB", resource.CategoryNetwork, "alb",
		resource.WithMetadata(map[string]string{
			"subnet_ids":         "subnet-456",
			"security_group_ids": "sg-789",
		}))
	endpoint, _ := resource.NewResource("vpce-678", "VPCEndpoint", resource.CategoryNetwork, "endpoint",
		resource.WithMetadata(map[string]string{"route_table_ids": "rtb-012"}))

	relationships, err := NewRelationshipBuilder().Build([]*resource.Resource{
		vpc, subnet, securityGroup, routeTable, alb, endpoint,
	})
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}

	wanted := map[string]relationship.RelationshipType{
		"alb-345->subnet-456": relationship.RelationshipAssociatedWith,
		"alb-345->sg-789":     relationship.RelationshipAssociatedWith,
		"vpce-678->rtb-012":   relationship.RelationshipAssociatedWith,
	}
	for _, rel := range relationships {
		key := rel.SourceID() + "->" + rel.TargetID()
		if expected, ok := wanted[key]; ok {
			if rel.Type() != expected {
				t.Errorf("%s type = %s, want %s", key, rel.Type(), expected)
			}
			delete(wanted, key)
		}
	}
	for key := range wanted {
		t.Errorf("missing attachment relationship %s", key)
	}
}

func TestRelationshipBuilder_LoadBalancerRelationships(t *testing.T) {
	lb, _ := resource.NewResource("lb-arn", "ALB", resource.CategoryNetwork, "lb")
	instance, _ := resource.NewResource("i-123", "EC2", resource.CategoryCompute, "instance")
	targetGroup, _ := resource.NewResource("tg-arn", "TargetGroup", resource.CategoryNetwork, "targets",
		resource.WithMetadata(map[string]string{"load_balancer_arn": "lb-arn", "instance_ids": "i-123"}))
	listener, _ := resource.NewResource("listener-arn", "Listener", resource.CategoryNetwork, "HTTPS:443",
		resource.WithMetadata(map[string]string{"default_target_group_arn": "tg-arn"}))

	relationships, err := NewRelationshipBuilder().Build([]*resource.Resource{lb, instance, targetGroup, listener})
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}

	wanted := map[string]relationship.RelationshipType{
		"tg-arn->lb-arn":       relationship.RelationshipBelongsTo,
		"tg-arn->i-123":        relationship.RelationshipTargets,
		"listener-arn->tg-arn": relationship.RelationshipTargets,
	}
	for _, rel := range relationships {
		key := rel.SourceID() + "->" + rel.TargetID()
		if expected, ok := wanted[key]; ok && rel.Type() == expected {
			delete(wanted, key)
		}
	}
	for key := range wanted {
		t.Errorf("missing load balancer relationship %s", key)
	}
}

func TestRelationshipBuilder_EC2NetworkInterfaceRelationships(t *testing.T) {
	instance, _ := resource.NewResource("i-123", "EC2", resource.CategoryCompute, "instance",
		resource.WithMetadata(map[string]string{"network_interface_ids": "eni-123"}))
	interfaceResource, _ := resource.NewResource("eni-123", "NetworkInterface", resource.CategoryNetwork, "interface",
		resource.WithMetadata(map[string]string{"instance_id": "i-123"}))

	relationships, err := NewRelationshipBuilder().Build([]*resource.Resource{instance, interfaceResource})
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}

	wanted := map[string]bool{"i-123->eni-123": false, "eni-123->i-123": false}
	for _, rel := range relationships {
		key := rel.SourceID() + "->" + rel.TargetID()
		if _, ok := wanted[key]; ok && rel.Type() == relationship.RelationshipAssociatedWith {
			wanted[key] = true
		}
	}
	for key, found := range wanted {
		if !found {
			t.Errorf("missing EC2/ENI relationship %s", key)
		}
	}
}

func TestRelationshipBuilder_ResolvesDNSAliasesOnce(t *testing.T) {
	distribution, _ := resource.NewResource("distribution-id", "CloudFrontDistribution", resource.CategoryNetwork, "distribution",
		resource.WithMetadata(map[string]string{"dns_name": "example.cloudfront.net"}))
	record, _ := resource.NewResource("zone:example.com:A", "Route53Record", resource.CategoryNetwork, "example.com",
		resource.WithMetadata(map[string]string{"alias_target": "example.cloudfront.net.,example.cloudfront.net"}))

	relationships, err := NewRelationshipBuilder().Build([]*resource.Resource{distribution, record})
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	if len(relationships) != 1 {
		t.Fatalf("Build() returned %d relationships, want 1", len(relationships))
	}
	rel := relationships[0]
	if rel.SourceID() != string(record.ID()) || rel.TargetID() != string(distribution.ID()) || rel.Type() != relationship.RelationshipTargets {
		t.Errorf("relationship = %s -> %s (%s), want %s -> %s (%s)", rel.SourceID(), rel.TargetID(), rel.Type(), record.ID(), distribution.ID(), relationship.RelationshipTargets)
	}
}
