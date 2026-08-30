package graph_test

import (
	"testing"

	"github.com/elip/WeaveLens/internal/domain/graph"
	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

func TestGraphAddNode(t *testing.T) {
	g := graph.NewGraph()

	node, err := resource.NewResource("res-1", "EC2", resource.CategoryCompute, "MyInstance")
	if err != nil {
		t.Fatalf("NewResource() error = %v", err)
	}

	err = g.AddNode(node)
	if err != nil {
		t.Errorf("AddNode() error = %v", err)
	}

	got, err := g.GetNode("res-1")
	if err != nil {
		t.Errorf("GetNode() error = %v", err)
	}

	if got.ID() != "res-1" {
		t.Errorf("GetNode() ID = %v, want res-1", got.ID())
	}
}

func TestGraphDuplicateNode(t *testing.T) {
	g := graph.NewGraph()

	node, err := resource.NewResource("res-1", "EC2", resource.CategoryCompute, "MyInstance")
	if err != nil {
		t.Fatalf("NewResource() error = %v", err)
	}

	err = g.AddNode(node)
	if err != nil {
		t.Errorf("AddNode() error = %v", err)
	}

	err = g.AddNode(node)
	if err != graph.ErrDuplicateNode {
		t.Errorf("AddNode() error = %v, want %v", err, graph.ErrDuplicateNode)
	}
}

func TestGraphAddRelationship(t *testing.T) {
	g := graph.NewGraph()

	rel, err := relationship.NewRelationship("rel-1", "vpc-1", "subnet-1", relationship.RelationshipContains)
	if err != nil {
		t.Fatalf("NewRelationship() error = %v", err)
	}

	err = g.AddRelationship(rel)
	if err != nil {
		t.Errorf("AddRelationship() error = %v", err)
	}

	got, err := g.GetRelationship("rel-1")
	if err != nil {
		t.Errorf("GetRelationship() error = %v", err)
	}

	if got.ID() != "rel-1" {
		t.Errorf("GetRelationship() ID = %v, want rel-1", got.ID())
	}
}

func TestGraphDuplicateRelationship(t *testing.T) {
	g := graph.NewGraph()

	rel, err := relationship.NewRelationship("rel-1", "vpc-1", "subnet-1", relationship.RelationshipContains)
	if err != nil {
		t.Fatalf("NewRelationship() error = %v", err)
	}

	err = g.AddRelationship(rel)
	if err != nil {
		t.Errorf("AddRelationship() error = %v", err)
	}

	err = g.AddRelationship(rel)
	if err != graph.ErrDuplicateRelationship {
		t.Errorf("AddRelationship() error = %v, want %v", err, graph.ErrDuplicateRelationship)
	}
}

func TestGraphNodeLookup(t *testing.T) {
	g := graph.NewGraph()

	_, err := g.GetNode("missing")
	if err != graph.ErrNodeNotFound {
		t.Errorf("GetNode() error = %v, want %v", err, graph.ErrNodeNotFound)
	}
}

func TestGraphRelationshipLookup(t *testing.T) {
	g := graph.NewGraph()

	_, err := g.GetRelationship("missing")
	if err != graph.ErrRelationshipNotFound {
		t.Errorf("GetRelationship() error = %v, want %v", err, graph.ErrRelationshipNotFound)
	}
}

func TestGraphCounts(t *testing.T) {
	g := graph.NewGraph()

	if g.NodeCount() != 0 {
		t.Errorf("NodeCount() = %v, want 0", g.NodeCount())
	}
	if g.RelationshipCount() != 0 {
		t.Errorf("RelationshipCount() = %v, want 0", g.RelationshipCount())
	}

	node, _ := resource.NewResource("res-1", "EC2", resource.CategoryCompute, "MyInstance")
	_ = g.AddNode(node)

	if g.NodeCount() != 1 {
		t.Errorf("NodeCount() = %v, want 1", g.NodeCount())
	}
}

func TestGraphGetRelationshipsForNode(t *testing.T) {
	g := graph.NewGraph()

	vpc, _ := resource.NewResource("vpc-1", "VPC", resource.CategoryNetwork, "vpc-1")
	subnet, _ := resource.NewResource("subnet-1", "Subnet", resource.CategoryNetwork, "subnet-1")
	_ = g.AddNode(vpc)
	_ = g.AddNode(subnet)

	rel, _ := relationship.NewRelationship("rel-1", "vpc-1", "subnet-1", relationship.RelationshipContains)
	_ = g.AddRelationship(rel)

	rels, err := g.GetRelationshipsForNode("vpc-1")
	if err != nil {
		t.Fatalf("GetRelationshipsForNode() error = %v", err)
	}

	if len(rels) != 1 {
		t.Errorf("GetRelationshipsForNode() length = %v, want 1", len(rels))
	}

	_, err = g.GetRelationshipsForNode("missing")
	if err != graph.ErrNodeNotFound {
		t.Errorf("GetRelationshipsForNode() error = %v, want %v", err, graph.ErrNodeNotFound)
	}
}

func TestGraphGetNeighbors(t *testing.T) {
	g := graph.NewGraph()

	vpc, _ := resource.NewResource("vpc-1", "VPC", resource.CategoryNetwork, "vpc-1")
	subnet, _ := resource.NewResource("subnet-1", "Subnet", resource.CategoryNetwork, "subnet-1")
	ec2, _ := resource.NewResource("ec2-1", "EC2", resource.CategoryCompute, "ec2-1")
	_ = g.AddNode(vpc)
	_ = g.AddNode(subnet)
	_ = g.AddNode(ec2)

	rel1, _ := relationship.NewRelationship("rel-1", "vpc-1", "subnet-1", relationship.RelationshipContains)
	rel2, _ := relationship.NewRelationship("rel-2", "subnet-1", "ec2-1", relationship.RelationshipContains)
	_ = g.AddRelationship(rel1)
	_ = g.AddRelationship(rel2)

	neighbors, err := g.GetNeighbors("subnet-1")
	if err != nil {
		t.Fatalf("GetNeighbors() error = %v", err)
	}

	if len(neighbors) != 2 {
		t.Errorf("GetNeighbors() length = %v, want 2", len(neighbors))
	}

	_, err = g.GetNeighbors("missing")
	if err != graph.ErrNodeNotFound {
		t.Errorf("GetNeighbors() error = %v, want %v", err, graph.ErrNodeNotFound)
	}
}

func TestGraphFilterNodes(t *testing.T) {
	g := graph.NewGraph()

	vpc, _ := resource.NewResource("vpc-1", "VPC", resource.CategoryNetwork, "vpc-1")
	ec2, _ := resource.NewResource("ec2-1", "EC2", resource.CategoryCompute, "ec2-1")
	_ = g.AddNode(vpc)
	_ = g.AddNode(ec2)

	networkNodes := g.FilterNodes(func(node *resource.Resource) bool {
		return node.Category() == resource.CategoryNetwork
	})

	if len(networkNodes) != 1 {
		t.Errorf("FilterNodes() length = %v, want 1", len(networkNodes))
	}

	if networkNodes[0].ID() != "vpc-1" {
		t.Errorf("FilterNodes()[0].ID() = %v, want vpc-1", networkNodes[0].ID())
	}
}

func TestGraphFilterRelationships(t *testing.T) {
	g := graph.NewGraph()

	rel1, _ := relationship.NewRelationship("rel-1", "vpc-1", "subnet-1", relationship.RelationshipContains)
	rel2, _ := relationship.NewRelationship("rel-2", "subnet-1", "ec2-1", relationship.RelationshipConnectsTo)
	_ = g.AddRelationship(rel1)
	_ = g.AddRelationship(rel2)

	containsRels := g.FilterRelationships(func(rel *relationship.Relationship) bool {
		return rel.Type() == relationship.RelationshipContains
	})

	if len(containsRels) != 1 {
		t.Errorf("FilterRelationships() length = %v, want 1", len(containsRels))
	}
}

func TestGraphEmptyGraph(t *testing.T) {
	g := graph.NewGraph()

	if g.NodeCount() != 0 {
		t.Errorf("NodeCount() = %v, want 0", g.NodeCount())
	}
	if g.RelationshipCount() != 0 {
		t.Errorf("RelationshipCount() = %v, want 0", g.RelationshipCount())
	}

	_, err := g.GetRelationshipsForNode("missing")
	if err != graph.ErrNodeNotFound {
		t.Errorf("GetRelationshipsForNode() error = %v, want %v", err, graph.ErrNodeNotFound)
	}

	_, err = g.GetNeighbors("missing")
	if err != graph.ErrNodeNotFound {
		t.Errorf("GetNeighbors() error = %v, want %v", err, graph.ErrNodeNotFound)
	}
}

func TestGraphSnapshot(t *testing.T) {
	g := graph.NewGraph()

	vpc, _ := resource.NewResource("vpc-1", "VPC", resource.CategoryNetwork, "vpc-1")
	subnet, _ := resource.NewResource("subnet-1", "Subnet", resource.CategoryNetwork, "subnet-1")
	_ = g.AddNode(vpc)
	_ = g.AddNode(subnet)

	rel, _ := relationship.NewRelationship("rel-1", "vpc-1", "subnet-1", relationship.RelationshipContains)
	_ = g.AddRelationship(rel)

	snapshot := g.Snapshot()

	if len(snapshot.Nodes) != 2 {
		t.Errorf("Snapshot.Nodes length = %v, want 2", len(snapshot.Nodes))
	}
	if len(snapshot.Relationships) != 1 {
		t.Errorf("Snapshot.Relationships length = %v, want 1", len(snapshot.Relationships))
	}

	if _, exists := snapshot.Nodes["vpc-1"]; !exists {
		t.Errorf("Snapshot.Nodes missing vpc-1")
	}
	if _, exists := snapshot.Relationships["rel-1"]; !exists {
		t.Errorf("Snapshot.Relationships missing rel-1")
	}
}

func TestGraphLargeGraph(t *testing.T) {
	g := graph.NewGraph()

	const nodeCount = 1000
	for i := 0; i < nodeCount; i++ {
		id := resource.ResourceID(string(rune('a' + i%26)) + string(rune('0' + i/26)))
		node, _ := resource.NewResource(id, "EC2", resource.CategoryCompute, "instance-"+string(rune('0'+i)))
		_ = g.AddNode(node)
	}

	if g.NodeCount() != nodeCount {
		t.Errorf("NodeCount() = %v, want %d", g.NodeCount(), nodeCount)
	}

	_ = g.FilterNodes(func(node *resource.Resource) bool {
		return node.Category() == resource.CategoryCompute
	})

	_ = g.Snapshot()
}
