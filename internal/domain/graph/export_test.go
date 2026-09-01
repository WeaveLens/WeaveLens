package graph_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/elip/WeaveLens/internal/domain/graph"
	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

func newTestResource(id, name, typ, category string) *resource.Resource {
	r, _ := resource.NewResource(
		resource.ResourceID(id),
		resource.ResourceType(typ),
		resource.ResourceCategory(category),
		name,
	)
	return r
}

func newTestRelationship(id, sourceID, targetID, typ string) *relationship.Relationship {
	r, _ := relationship.NewRelationship(
		relationship.RelationshipID(id),
		sourceID,
		targetID,
		relationship.RelationshipType(typ),
	)
	return r
}

func TestNewExportGraph_EmptyGraph(t *testing.T) {
	g := graph.NewGraph()
	exportGraph := graph.NewExportGraph(g.Snapshot())

	if exportGraph.NodeCount != 0 {
		t.Errorf("Expected 0 nodes, got %d", exportGraph.NodeCount)
	}
	if exportGraph.EdgeCount != 0 {
		t.Errorf("Expected 0 edges, got %d", exportGraph.EdgeCount)
	}
	if len(exportGraph.Nodes) != 0 {
		t.Errorf("Expected 0 nodes in slice, got %d", len(exportGraph.Nodes))
	}
	if len(exportGraph.Edges) != 0 {
		t.Errorf("Expected 0 edges in slice, got %d", len(exportGraph.Edges))
	}
}

func TestNewExportGraph_SimpleGraph(t *testing.T) {
	g := graph.NewGraph()

	vpc := newTestResource("vpc-123", "test-vpc", "VPC", "network")
	ec2 := newTestResource("i-456", "test-ec2", "EC2", "compute")

	_ = g.AddNode(vpc)
	_ = g.AddNode(ec2)

	rel := newTestRelationship("rel-1", "vpc-123", "i-456", "contains")
	_ = g.AddRelationship(rel)

	exportGraph := graph.NewExportGraph(g.Snapshot())

	if exportGraph.NodeCount != 2 {
		t.Errorf("Expected 2 nodes, got %d", exportGraph.NodeCount)
	}
	if exportGraph.EdgeCount != 1 {
		t.Errorf("Expected 1 edge, got %d", exportGraph.EdgeCount)
	}
}

func TestExportGraph_ToJSON(t *testing.T) {
	g := graph.NewGraph()

	vpc := newTestResource("vpc-123", "test-vpc", "VPC", "network")
	_ = g.AddNode(vpc)

	exportGraph := graph.NewExportGraph(g.Snapshot())
	data, err := exportGraph.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if result["nodeCount"].(float64) != 1 {
		t.Errorf("Expected nodeCount 1, got %v", result["nodeCount"])
	}

	nodes := result["nodes"].([]interface{})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	node := nodes[0].(map[string]interface{})
	if node["id"] != "vpc-123" {
		t.Errorf("Expected id vpc-123, got %v", node["id"])
	}
	if node["name"] != "test-vpc" {
		t.Errorf("Expected name test-vpc, got %v", node["name"])
	}
	if node["type"] != "VPC" {
		t.Errorf("Expected type VPC, got %v", node["type"])
	}
	if node["category"] != "network" {
		t.Errorf("Expected category network, got %v", node["category"])
	}
}

func TestExportGraph_ToJSON_DuplicateResources(t *testing.T) {
	g := graph.NewGraph()

	vpc1 := newTestResource("vpc-123", "test-vpc", "VPC", "network")
	vpc2 := newTestResource("vpc-123", "test-vpc", "VPC", "network")

	err := g.AddNode(vpc1)
	if err != nil {
		t.Fatalf("First add should succeed: %v", err)
	}
	err = g.AddNode(vpc2)
	if err == nil {
		t.Fatal("Second add should fail for duplicate")
	}

	exportGraph := graph.NewExportGraph(g.Snapshot())
	if exportGraph.NodeCount != 1 {
		t.Errorf("Expected 1 node (duplicate rejected), got %d", exportGraph.NodeCount)
	}
}

func TestExportGraph_ToDrawIO(t *testing.T) {
	g := graph.NewGraph()

	vpc := newTestResource("vpc-123", "test-vpc", "VPC", "network")
	ec2 := newTestResource("i-456", "test-ec2", "EC2", "compute")
	_ = g.AddNode(vpc)
	_ = g.AddNode(ec2)

	rel := newTestRelationship("rel-1", "vpc-123", "i-456", "contains")
	_ = g.AddRelationship(rel)

	exportGraph := graph.NewExportGraph(g.Snapshot())
	data, err := exportGraph.ToDrawIO()
	if err != nil {
		t.Fatalf("ToDrawIO() error: %v", err)
	}

	xmlContent := string(data)

	if !strings.Contains(xmlContent, "<mxfile>") {
		t.Error("Expected mxfile element")
	}
	if !strings.Contains(xmlContent, "<mxGraphModel>") {
		t.Error("Expected mxGraphModel element")
	}
	if !strings.Contains(xmlContent, "test-vpc") {
		t.Error("Expected node label in output")
	}
	if !strings.Contains(xmlContent, "contains") {
		t.Error("Expected edge label in output")
	}
}

func TestExportGraph_ToDrawIO_EmptyGraph(t *testing.T) {
	g := graph.NewGraph()
	exportGraph := graph.NewExportGraph(g.Snapshot())

	data, err := exportGraph.ToDrawIO()
	if err != nil {
		t.Fatalf("ToDrawIO() error: %v", err)
	}

	xmlContent := string(data)
	if !strings.Contains(xmlContent, "<mxfile>") {
		t.Error("Expected mxfile element even for empty graph")
	}
}

func TestExportGraph_ToSVG(t *testing.T) {
	g := graph.NewGraph()

	vpc := newTestResource("vpc-123", "test-vpc", "VPC", "network")
	ec2 := newTestResource("i-456", "test-ec2", "EC2", "compute")
	_ = g.AddNode(vpc)
	_ = g.AddNode(ec2)

	rel := newTestRelationship("rel-1", "vpc-123", "i-456", "contains")
	_ = g.AddRelationship(rel)

	exportGraph := graph.NewExportGraph(g.Snapshot())
	data, err := exportGraph.ToSVG()
	if err != nil {
		t.Fatalf("ToSVG() error: %v", err)
	}

	svgContent := string(data)

	if !strings.Contains(svgContent, "<svg") {
		t.Error("Expected svg element")
	}
	if !strings.Contains(svgContent, "test-vpc") {
		t.Error("Expected node label in output")
	}
	if !strings.Contains(svgContent, "test-ec2") {
		t.Error("Expected node label in output")
	}
	if !strings.Contains(svgContent, "contains") {
		t.Error("Expected edge label in output")
	}
}

func TestExportGraph_ToSVG_EmptyGraph(t *testing.T) {
	g := graph.NewGraph()
	exportGraph := graph.NewExportGraph(g.Snapshot())

	data, err := exportGraph.ToSVG()
	if err != nil {
		t.Fatalf("ToSVG() error: %v", err)
	}

	svgContent := string(data)
	if !strings.Contains(svgContent, "<svg") {
		t.Error("Expected svg element even for empty graph")
	}
}

func TestExportGraph_ComplexGraph(t *testing.T) {
	g := graph.NewGraph()

	vpc := newTestResource("vpc-123", "main-vpc", "VPC", "network")
	subnet := newTestResource("subnet-456", "private-subnet", "Subnet", "network")
	ec2a := newTestResource("i-001", "web-server-1", "EC2", "compute")
	ec2b := newTestResource("i-002", "web-server-2", "EC2", "compute")
	rds := newTestResource("db-789", "main-db", "RDS", "database")
	sg := newTestResource("sg-101", "web-sg", "SecurityGroup", "security")

	_ = g.AddNode(vpc)
	_ = g.AddNode(subnet)
	_ = g.AddNode(ec2a)
	_ = g.AddNode(ec2b)
	_ = g.AddNode(rds)
	_ = g.AddNode(sg)

	_ = g.AddRelationship(newTestRelationship("r1", "vpc-123", "subnet-456", "contains"))
	_ = g.AddRelationship(newTestRelationship("r2", "subnet-456", "i-001", "contains"))
	_ = g.AddRelationship(newTestRelationship("r3", "subnet-456", "i-002", "contains"))
	_ = g.AddRelationship(newTestRelationship("r4", "subnet-456", "db-789", "contains"))
	_ = g.AddRelationship(newTestRelationship("r5", "sg-101", "i-001", "associated_with"))
	_ = g.AddRelationship(newTestRelationship("r6", "sg-101", "i-002", "associated_with"))

	exportGraph := graph.NewExportGraph(g.Snapshot())

	if exportGraph.NodeCount != 6 {
		t.Errorf("Expected 6 nodes, got %d", exportGraph.NodeCount)
	}
	if exportGraph.EdgeCount != 6 {
		t.Errorf("Expected 6 edges, got %d", exportGraph.EdgeCount)
	}

	jsonData, err := exportGraph.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if result["nodeCount"].(float64) != 6 {
		t.Errorf("Expected nodeCount 6 in JSON, got %v", result["nodeCount"])
	}
	if result["edgeCount"].(float64) != 6 {
		t.Errorf("Expected edgeCount 6 in JSON, got %v", result["edgeCount"])
	}
}

func TestExportGraph_RelationshipRendering(t *testing.T) {
	g := graph.NewGraph()

	a := newTestResource("a", "NodeA", "TypeA", "network")
	b := newTestResource("b", "NodeB", "TypeB", "compute")
	_ = g.AddNode(a)
	_ = g.AddNode(b)

	rel := newTestRelationship("rel-ab", "a", "b", "connects_to")
	_ = g.AddRelationship(rel)

	exportGraph := graph.NewExportGraph(g.Snapshot())

	drawioData, err := exportGraph.ToDrawIO()
	if err != nil {
		t.Fatalf("ToDrawIO() error: %v", err)
	}
	if !strings.Contains(string(drawioData), "connects_to") {
		t.Error("Expected relationship type in Draw.io output")
	}

	svgData, err := exportGraph.ToSVG()
	if err != nil {
		t.Fatalf("ToSVG() error: %v", err)
	}
	if !strings.Contains(string(svgData), "connects_to") {
		t.Error("Expected relationship type in SVG output")
	}
}
