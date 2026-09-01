package graph

import (
	"encoding/json"
	"time"
)

type ExportNode struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Category string            `json:"category"`
	ARN      string            `json:"arn,omitempty"`
	Region   string            `json:"region,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type ExportEdge struct {
	ID       string            `json:"id"`
	SourceID string            `json:"sourceId"`
	TargetID string            `json:"targetId"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type ExportGraph struct {
	ExportedAt    time.Time     `json:"exportedAt"`
	NodeCount     int           `json:"nodeCount"`
	EdgeCount     int           `json:"edgeCount"`
	Nodes         []ExportNode  `json:"nodes"`
	Edges         []ExportEdge  `json:"edges"`
}

func NewExportGraph(snapshot GraphSnapshot) *ExportGraph {
	nodes := make([]ExportNode, 0, len(snapshot.Nodes))
	for _, node := range snapshot.Nodes {
		nodes = append(nodes, ExportNode{
			ID:       string(node.ID()),
			Name:     node.Name(),
			Type:     string(node.Type()),
			Category: string(node.Category()),
			ARN:      node.ARN(),
			Region:   node.Region(),
			Metadata: node.Metadata(),
		})
	}

	edges := make([]ExportEdge, 0, len(snapshot.Relationships))
	for _, rel := range snapshot.Relationships {
		edges = append(edges, ExportEdge{
			ID:       string(rel.ID()),
			SourceID: rel.SourceID(),
			TargetID: rel.TargetID(),
			Type:     string(rel.Type()),
			Metadata: rel.Metadata(),
		})
	}

	return &ExportGraph{
		ExportedAt: time.Now().UTC(),
		NodeCount:  len(nodes),
		EdgeCount:  len(edges),
		Nodes:      nodes,
		Edges:      edges,
	}
}

func (g *ExportGraph) ToJSON() ([]byte, error) {
	return json.MarshalIndent(g, "", "  ")
}

func (g *ExportGraph) ToCompactJSON() ([]byte, error) {
	return json.Marshal(g)
}
