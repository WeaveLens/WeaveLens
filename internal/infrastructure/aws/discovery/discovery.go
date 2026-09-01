package discovery

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/discovery/resilience"
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

func (s *Service) Discover(ctx context.Context, request DiscoveryRequest) (*DiscoveryResult, error) {
	result := &DiscoveryResult{
		Resources:     make([]*resource.Resource, 0),
		Relationships: make([]*relationship.Relationship, 0),
		Errors:        make([]error, 0),
	}

	resources, errs := s.scanConcurrently(ctx)
	for _, res := range resources {
		if request.Region == "" || res.Region() == request.Region {
			result.AddResource(res)
		}
	}
	for _, err := range errs {
		result.AddError(err)
	}

	if ctx.Err() != nil {
		result.AddError(fmt.Errorf("scan cancelled: %w", ctx.Err()))
		return result, ctx.Err()
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

func (s *Service) scanConcurrently(ctx context.Context) ([]*resource.Resource, []error) {
	workers := s.config.Workers
	if workers > len(s.scanners) {
		workers = len(s.scanners)
	}

	resources := make([][]*resource.Resource, len(s.scanners))
	scanErrors := make([]error, len(s.scanners))

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, scanner := range s.scanners {
		i, scanner := i, scanner
		wg.Add(1)

		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-cancelCtx.Done():
				mu.Lock()
				scanErrors[i] = fmt.Errorf("scanner %s: %w", scanner.Name(), cancelCtx.Err())
				mu.Unlock()
				return
			}

			if cancelCtx.Err() != nil {
				mu.Lock()
				scanErrors[i] = fmt.Errorf("scanner %s: %w", scanner.Name(), cancelCtx.Err())
				mu.Unlock()
				return
			}

			scanFunc := func() error {
				res, err := scanner.Scan(cancelCtx)
				if err != nil {
					return err
				}
				mu.Lock()
				resources[i] = res
				mu.Unlock()
				return nil
			}

			var err error
			if s.config.RetryConfig.MaxAttempts > 0 {
				isRetryable := func(err error) bool {
					return isRetryableError(err)
				}
				err = resilience.Retry(cancelCtx, s.config.RetryConfig, isRetryable, scanFunc)
			} else {
				err = scanFunc()
			}

			if err != nil {
				mu.Lock()
				scanErrors[i] = &ScannerError{Scanner: scanner.Name(), Err: err}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if ctx.Err() != nil {
		for i := range scanErrors {
			if scanErrors[i] == nil {
				scanErrors[i] = fmt.Errorf("scanner %s: %w", s.scanners[i].Name(), ctx.Err())
			}
		}
	}

	var allResources []*resource.Resource
	var allErrors []error

	for _, res := range resources {
		if res != nil {
			allResources = append(allResources, res...)
		}
	}
	for _, err := range scanErrors {
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	return allResources, allErrors
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrThrottling) {
		return true
	}
	if errors.Is(err, ErrTransientFailure) {
		return true
	}
	if errors.Is(err, ErrContextCanceled) {
		return false
	}
	if errors.Is(err, ErrAccessDenied) {
		return false
	}

	return false
}
