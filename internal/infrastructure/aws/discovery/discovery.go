package discovery

import (
	"context"
	"fmt"

	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type DiscoveryRequest struct {
	Region string
}

type DiscoveryResult struct {
	Resources     []*resource.Resource
	Relationships []*relationship.Relationship
	Errors        []error
}

func (r *DiscoveryResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *DiscoveryResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
}

func (r *DiscoveryResult) AddResource(res *resource.Resource) {
	r.Resources = append(r.Resources, res)
}

func (r *DiscoveryResult) AddRelationship(rel *relationship.Relationship) {
	r.Relationships = append(r.Relationships, rel)
}

type ResourceDiscovery interface {
	Discover(ctx context.Context, request DiscoveryRequest) (*DiscoveryResult, error)
}

type Scanner interface {
	Scan(ctx context.Context) ([]*resource.Resource, error)
}

type RelationshipBuilder interface {
	Build(resources []*resource.Resource) ([]*relationship.Relationship, error)
}

type Service struct {
	scanners    []Scanner
	relationshipBuilder RelationshipBuilder
}

func NewService(scanners []Scanner, rb RelationshipBuilder) *Service {
	return &Service{
		scanners:    scanners,
		relationshipBuilder: rb,
	}
}

func (s *Service) Discover(ctx context.Context, request DiscoveryRequest) (*DiscoveryResult, error) {
	result := &DiscoveryResult{
		Resources:     make([]*resource.Resource, 0),
		Relationships: make([]*relationship.Relationship, 0),
		Errors:        make([]error, 0),
	}

	for _, scanner := range s.scanners {
		select {
		case <-ctx.Done():
			result.AddError(fmt.Errorf("scan cancelled: %w", ctx.Err()))
			return result, ctx.Err()
		default:
		}

		resources, err := scanner.Scan(ctx)
		if err != nil {
			result.AddError(err)
			continue
		}
		for _, res := range resources {
			result.AddResource(res)
		}
	}

	relationships, err := s.relationshipBuilder.Build(result.Resources)
	if err != nil {
		result.AddError(err)
	} else {
		for _, rel := range relationships {
			result.AddRelationship(rel)
		}
	}

	return result, nil
}
