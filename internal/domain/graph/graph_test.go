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
