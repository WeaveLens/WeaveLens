package graph

import (
	"errors"
	"sync"

	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

var (
	ErrNodeNotFound            = errors.New("node not found")
	ErrDuplicateNode           = errors.New("duplicate node")
	ErrRelationshipNotFound    = errors.New("relationship not found")
	ErrDuplicateRelationship   = errors.New("duplicate relationship")
)

type Graph struct {
	nodes map[resource.ResourceID]*resource.Resource
	edges map[relationship.RelationshipID]*relationship.Relationship
	mu    sync.RWMutex
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[resource.ResourceID]*resource.Resource),
		edges: make(map[relationship.RelationshipID]*relationship.Relationship),
	}
}

func (g *Graph) AddNode(node *resource.Resource) error {
	if node == nil {
		return errors.New("node must not be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[node.ID()]; exists {
		return ErrDuplicateNode
	}

	g.nodes[node.ID()] = node
	return nil
}

func (g *Graph) AddRelationship(rel *relationship.Relationship) error {
	if rel == nil {
		return errors.New("relationship must not be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.edges[rel.ID()]; exists {
		return ErrDuplicateRelationship
	}

	g.edges[rel.ID()] = rel
	return nil
}

func (g *Graph) GetNode(id resource.ResourceID) (*resource.Resource, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[id]
	if !exists {
		return nil, ErrNodeNotFound
	}

	return node, nil
}

func (g *Graph) GetRelationship(id relationship.RelationshipID) (*relationship.Relationship, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	rel, exists := g.edges[id]
	if !exists {
		return nil, ErrRelationshipNotFound
	}

	return rel, nil
}

func (g *Graph) Nodes() map[resource.ResourceID]*resource.Resource {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make(map[resource.ResourceID]*resource.Resource, len(g.nodes))
	for id, node := range g.nodes {
		result[id] = node
	}

	return result
}

func (g *Graph) Relationships() map[relationship.RelationshipID]*relationship.Relationship {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make(map[relationship.RelationshipID]*relationship.Relationship, len(g.edges))
	for id, rel := range g.edges {
		result[id] = rel
	}

	return result
}

func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.nodes)
}

func (g *Graph) RelationshipCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.edges)
}

func (g *Graph) GetRelationshipsForNode(id resource.ResourceID) ([]*relationship.Relationship, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[id]; !exists {
		return nil, ErrNodeNotFound
	}

	var result []*relationship.Relationship
	for _, rel := range g.edges {
		if rel.SourceID() == string(id) || rel.TargetID() == string(id) {
			result = append(result, rel)
		}
	}

	return result, nil
}

func (g *Graph) GetNeighbors(id resource.ResourceID) ([]*resource.Resource, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[id]; !exists {
		return nil, ErrNodeNotFound
	}

	neighborIDs := make(map[resource.ResourceID]struct{})
	for _, rel := range g.edges {
		if rel.SourceID() == string(id) {
			neighborIDs[resource.ResourceID(rel.TargetID())] = struct{}{}
		}
		if rel.TargetID() == string(id) {
			neighborIDs[resource.ResourceID(rel.SourceID())] = struct{}{}
		}
	}

	var result []*resource.Resource
	for neighborID := range neighborIDs {
		if node, exists := g.nodes[neighborID]; exists {
			result = append(result, node)
		}
	}

	return result, nil
}

func (g *Graph) FilterNodes(predicate func(*resource.Resource) bool) []*resource.Resource {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []*resource.Resource
	for _, node := range g.nodes {
		if predicate(node) {
			result = append(result, node)
		}
	}

	return result
}

func (g *Graph) FilterRelationships(predicate func(*relationship.Relationship) bool) []*relationship.Relationship {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []*relationship.Relationship
	for _, rel := range g.edges {
		if predicate(rel) {
			result = append(result, rel)
		}
	}

	return result
}

type GraphSnapshot struct {
	Nodes        map[resource.ResourceID]*resource.Resource
	Relationships map[relationship.RelationshipID]*relationship.Relationship
}

func (g *Graph) Snapshot() GraphSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make(map[resource.ResourceID]*resource.Resource, len(g.nodes))
	for id, node := range g.nodes {
		nodes[id] = node
	}

	rels := make(map[relationship.RelationshipID]*relationship.Relationship, len(g.edges))
	for id, rel := range g.edges {
		rels[id] = rel
	}

	return GraphSnapshot{
		Nodes:        nodes,
		Relationships: rels,
	}
}
