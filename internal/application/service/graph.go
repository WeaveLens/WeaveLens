package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/elip/WeaveLens/internal/domain/graph"
	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/discovery"
	"github.com/elip/WeaveLens/internal/infrastructure/nats"
)

type graphService struct {
	eventBus  *nats.EventBus
	logger    *slog.Logger
	mu        sync.RWMutex
	graphs    map[string]*graph.Graph
	discovery discovery.ResourceDiscovery
	onScanComplete func(scanID string, nodeCount, edgeCount int)
	scanRegions map[string]string
	history *ScanHistory
}

func NewGraphService(eventBus *nats.EventBus, logger *slog.Logger, discovery discovery.ResourceDiscovery) GraphService {
	return &graphService{
		eventBus:  eventBus,
		logger:    logger,
		graphs:    make(map[string]*graph.Graph),
		discovery: discovery,
		scanRegions: make(map[string]string),
	}
}

func NewGraphServiceWithCallback(eventBus *nats.EventBus, logger *slog.Logger, discovery discovery.ResourceDiscovery, onScanComplete func(scanID string, nodeCount, edgeCount int)) GraphService {
	return &graphService{
		eventBus:  eventBus,
		logger:    logger,
		graphs:    make(map[string]*graph.Graph),
		discovery: discovery,
		onScanComplete: onScanComplete,
		scanRegions: make(map[string]string),
	}
}

func (s *graphService) SetScanRegion(scanID, region string) {
	s.mu.Lock()
	s.scanRegions[scanID] = region
	s.mu.Unlock()
}

func (s *graphService) SetHistory(h *ScanHistory) {
	s.mu.Lock()
	s.history = h
	s.mu.Unlock()
}

func (s *graphService) buildGraph(scanID string) (*graph.Graph, error) {
	s.mu.RLock()
	g, exists := s.graphs[scanID]
	s.mu.RUnlock()

	if exists {
		return g, nil
	}

	if s.discovery == nil {
		return graph.NewGraph(), nil
	}

	if s.history != nil {
		if graphData, found := s.history.GetGraph(scanID); found {
			g = graph.NewGraph()
			for _, res := range graphData.Nodes {
				r, err := resource.NewResource(
					resource.ResourceID(res.ID),
					resource.ResourceType(res.Type),
					resource.ResourceCategory(res.Category),
					res.Name,
					resource.WithARN(res.ARN),
					resource.WithRegion(res.Region),
					resource.WithMetadata(res.Metadata),
					resource.WithTags(res.Tags),
				)
				if err != nil {
					continue
				}
				_ = g.AddNode(r)
			}
			for _, rel := range graphData.Edges {
				r, err := relationship.NewRelationship(
					relationship.RelationshipID(rel.ID),
					rel.SourceID,
					rel.TargetID,
					relationship.RelationshipType(rel.Type),
					relationship.WithRelationshipMetadata(rel.Metadata),
				)
				if err != nil {
					continue
				}
				_ = g.AddRelationship(r)
			}
			s.mu.Lock()
			s.graphs[scanID] = g
			s.mu.Unlock()
			return g, nil
		}
	}

	region := ""
	if scanID != "" {
		s.mu.RLock()
		region = s.scanRegions[scanID]
		s.mu.RUnlock()
	}

	ctx := context.Background()
	result, err := s.discovery.Discover(ctx, discovery.DiscoveryRequest{Region: region})
	if err != nil {
		return nil, fmt.Errorf("failed to discover resources: %w", err)
	}

	g = graph.NewGraph()
	for _, res := range result.Resources {
		if err := g.AddNode(res); err != nil && err != graph.ErrDuplicateNode {
			s.logger.Warn("failed to add node to graph", "error", err, "resourceID", res.ID())
		}
	}

	for _, rel := range result.Relationships {
		if err := g.AddRelationship(rel); err != nil && err != graph.ErrDuplicateRelationship {
			s.logger.Warn("failed to add relationship to graph", "error", err, "relationshipID", rel.ID())
		}
	}

	s.mu.Lock()
	s.graphs[scanID] = g
	s.mu.Unlock()

	if s.onScanComplete != nil && scanID != "" {
		s.onScanComplete(scanID, g.NodeCount(), g.RelationshipCount())
	}

	if s.history != nil && scanID != "" {
		nodes := g.Nodes()
		relationships := g.Relationships()
		var resources []Resource
		for _, node := range nodes {
			resources = append(resources, Resource{
				ID:       string(node.ID()),
				Name:     node.Name(),
				Type:     string(node.Type()),
				Category: string(node.Category()),
				ARN:      node.ARN(),
				Region:   node.Region(),
				Metadata: node.Metadata(),
				Tags:     node.Tags(),
			})
		}
		var rels []Relationship
		for _, rel := range relationships {
			rels = append(rels, Relationship{
				ID:       string(rel.ID()),
				SourceID: rel.SourceID(),
				TargetID: rel.TargetID(),
				Type:     string(rel.Type()),
				Metadata: rel.Metadata(),
			})
		}
		s.history.SaveGraph(scanID, resources, rels)
	}

	return g, nil
}

func (s *graphService) GetGraph(ctx context.Context, scanID string) ([]Resource, []Relationship, error) {
	g, err := s.buildGraph(scanID)
	if err != nil {
		return nil, nil, err
	}

	nodes := g.Nodes()
	relationships := g.Relationships()

	var resources []Resource
	for _, node := range nodes {
		resources = append(resources, Resource{
			ID:       string(node.ID()),
			Name:     node.Name(),
			Type:     string(node.Type()),
			Category: string(node.Category()),
			ARN:      node.ARN(),
			Region:   node.Region(),
			Metadata: node.Metadata(),
			Tags:     node.Tags(),
		})
	}

	var rels []Relationship
	for _, rel := range relationships {
		rels = append(rels, Relationship{
			ID:       string(rel.ID()),
			SourceID: rel.SourceID(),
			TargetID: rel.TargetID(),
			Type:     string(rel.Type()),
			Metadata: rel.Metadata(),
		})
	}

	return resources, rels, nil
}

func (s *graphService) GetResource(ctx context.Context, resourceID string) (Resource, error) {
	g, err := s.buildGraph("")
	if err != nil {
		return Resource{}, err
	}

	node, err := g.GetNode(resource.ResourceID(resourceID))
	if err != nil {
		return Resource{}, err
	}

	return Resource{
		ID:       string(node.ID()),
		Name:     node.Name(),
		Type:     string(node.Type()),
		Category: string(node.Category()),
		ARN:      node.ARN(),
		Region:   node.Region(),
		Metadata: node.Metadata(),
		Tags:     node.Tags(),
	}, nil
}

func (s *graphService) GetNeighbors(ctx context.Context, resourceID string) ([]Resource, error) {
	g, err := s.buildGraph("")
	if err != nil {
		return nil, err
	}

	neighbors, err := g.GetNeighbors(resource.ResourceID(resourceID))
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for _, node := range neighbors {
		resources = append(resources, Resource{
			ID:       string(node.ID()),
			Name:     node.Name(),
			Type:     string(node.Type()),
			Category: string(node.Category()),
			ARN:      node.ARN(),
			Region:   node.Region(),
			Metadata: node.Metadata(),
			Tags:     node.Tags(),
		})
	}
	return resources, nil
}

func (s *graphService) GetRelationships(ctx context.Context, resourceID string) ([]Relationship, error) {
	g, err := s.buildGraph("")
	if err != nil {
		return nil, err
	}

	rels, err := g.GetRelationshipsForNode(resource.ResourceID(resourceID))
	if err != nil {
		return nil, err
	}

	var relationships []Relationship
	for _, rel := range rels {
		relationships = append(relationships, Relationship{
			ID:       string(rel.ID()),
			SourceID: rel.SourceID(),
			TargetID: rel.TargetID(),
			Type:     string(rel.Type()),
			Metadata: rel.Metadata(),
		})
	}
	return relationships, nil
}
