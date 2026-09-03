package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/discovery/resilience"
)

type DiscoveryRequest struct {
	Region  string
	Regions []string
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
	Name() string
	Scan(ctx context.Context) ([]*resource.Resource, error)
}

type RelationshipBuilder interface {
	Build(resources []*resource.Resource) ([]*relationship.Relationship, error)
}

type ServiceConfig struct {
	Workers     int
	RetryConfig resilience.RetryConfig
	RateLimiter *resilience.RateLimiter
}

func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Workers:     4,
		RetryConfig: resilience.DefaultRetryConfig(),
	}
}

type Service struct {
	scanners            []Scanner
	awsConfig           aws.Config
	factory             ScannerFactory
	relationshipBuilder RelationshipBuilder
	config              ServiceConfig
}

func NewService(scanners []Scanner, rb RelationshipBuilder) *Service {
	return NewServiceWithConfig(scanners, rb, DefaultServiceConfig())
}

func NewServiceWithConfig(scanners []Scanner, rb RelationshipBuilder, config ServiceConfig) *Service {
	return &Service{
		scanners:            scanners,
		relationshipBuilder: rb,
		config:              config,
	}
}

func NewServiceDynamic(awsCfg aws.Config, factory ScannerFactory, rb RelationshipBuilder, config ServiceConfig) *Service {
	return &Service{
		awsConfig:           awsCfg,
		factory:             factory,
		relationshipBuilder: rb,
		config:              config,
	}
}

type ScannerFactory func(region string) ([]Scanner, error)

func (s *Service) Discover(ctx context.Context, request DiscoveryRequest) (*DiscoveryResult, error) {
	result := &DiscoveryResult{
		Resources:     make([]*resource.Resource, 0),
		Relationships: make([]*relationship.Relationship, 0),
		Errors:        make([]error, 0),
	}

	scanners, err := s.getScanners(ctx, request)
	if err != nil {
		return result, err
	}

	allResources := make([]*resource.Resource, 0)
	for _, scanner := range scanners {
		resources, scanErr := scanner.Scan(ctx)
		if scanErr != nil {
			result.AddError(&ScannerError{Scanner: scanner.Name(), Err: scanErr})
			continue
		}
		for _, res := range resources {
			result.AddResource(res)
			allResources = append(allResources, res)
		}
	}

	if ctx.Err() != nil {
		result.AddError(fmt.Errorf("scan cancelled: %w", ctx.Err()))
		return result, ctx.Err()
	}

	relationships, err := s.relationshipBuilder.Build(allResources)
	if err != nil {
		result.AddError(err)
	} else {
		for _, rel := range relationships {
			result.AddRelationship(rel)
		}
	}

	return result, nil
}

func (s *Service) getScanners(ctx context.Context, request DiscoveryRequest) ([]Scanner, error) {
	if s.factory == nil {
		return s.scanners, nil
	}

	if request.Region != "" {
		return s.factory(request.Region)
	}

	if len(request.Regions) > 0 {
		var scanners []Scanner
		for _, r := range request.Regions {
			rs, err := s.factory(r)
			if err != nil {
				continue
			}
			scanners = append(scanners, rs...)
		}
		return scanners, nil
	}

	regions, err := availableRegions(ctx, s.awsConfig)
	if err != nil || len(regions) == 0 {
		regions = []string{"us-east-1"}
	}

	var scanners []Scanner
	for _, r := range regions {
		rs, err := s.factory(r)
		if err != nil {
			continue
		}
		scanners = append(scanners, rs...)
	}
	return scanners, nil
}
