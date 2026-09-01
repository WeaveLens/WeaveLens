package service

import (
	"context"
	"sync"
	"time"
)

type RegionInfo struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type RegionService struct {
	mu      sync.RWMutex
	regions []RegionInfo
	client  RegionFetcher
}

type RegionFetcher interface {
	FetchRegions(ctx context.Context) ([]RegionInfo, error)
}

func NewRegionService(client RegionFetcher) *RegionService {
	return &RegionService{
		client: client,
	}
}

func (s *RegionService) GetRegions(ctx context.Context) []RegionInfo {
	s.mu.RLock()
	if s.regions != nil {
		defer s.mu.RUnlock()
		return s.regions
	}
	s.mu.RUnlock()

	if s.client != nil {
		regions, err := s.client.FetchRegions(ctx)
		if err == nil && len(regions) > 0 {
			s.mu.Lock()
			s.regions = regions
			s.mu.Unlock()
			return regions
		}
	}

	return s.fallbackRegions()
}

func (s *RegionService) RefreshRegions(ctx context.Context) []RegionInfo {
	if s.client != nil {
		regions, err := s.client.FetchRegions(ctx)
		if err == nil && len(regions) > 0 {
			s.mu.Lock()
			s.regions = regions
			s.mu.Unlock()
			return regions
		}
	}
	return s.fallbackRegions()
}

func (s *RegionService) fallbackRegions() []RegionInfo {
	return []RegionInfo{
		{Value: "us-east-1", Label: "US East (N. Virginia)"},
		{Value: "us-east-2", Label: "US East (Ohio)"},
		{Value: "us-west-1", Label: "US West (N. California)"},
		{Value: "us-west-2", Label: "US West (Oregon)"},
		{Value: "ca-central-1", Label: "Canada (Central)"},
		{Value: "sa-east-1", Label: "South America (São Paulo)"},
		{Value: "eu-west-1", Label: "Europe (Ireland)"},
		{Value: "eu-west-2", Label: "Europe (London)"},
		{Value: "eu-west-3", Label: "Europe (Paris)"},
		{Value: "eu-central-1", Label: "Europe (Frankfurt)"},
		{Value: "eu-south-1", Label: "Europe (Milan)"},
		{Value: "eu-north-1", Label: "Europe (Stockholm)"},
		{Value: "me-south-1", Label: "Middle East (Bahrain)"},
		{Value: "af-south-1", Label: "Africa (Cape Town)"},
		{Value: "ap-southeast-1", Label: "Asia Pacific (Singapore)"},
		{Value: "ap-southeast-2", Label: "Asia Pacific (Sydney)"},
		{Value: "ap-south-1", Label: "Asia Pacific (Mumbai)"},
		{Value: "ap-northeast-1", Label: "Asia Pacific (Tokyo)"},
		{Value: "ap-northeast-2", Label: "Asia Pacific (Seoul)"},
		{Value: "ap-east-1", Label: "Asia Pacific (Hong Kong)"},
	}
}

var _ = time.Now
