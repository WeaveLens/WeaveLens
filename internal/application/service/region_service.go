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

var regionCatalog = []RegionInfo{
	{Value: "af-south-1", Label: "Africa (Cape Town)"},
	{Value: "ap-east-1", Label: "Asia Pacific (Hong Kong)"},
	{Value: "ap-east-2", Label: "Asia Pacific (Taipei)"},
	{Value: "ap-northeast-1", Label: "Asia Pacific (Tokyo)"},
	{Value: "ap-northeast-2", Label: "Asia Pacific (Seoul)"},
	{Value: "ap-northeast-3", Label: "Asia Pacific (Osaka)"},
	{Value: "ap-south-1", Label: "Asia Pacific (Mumbai)"},
	{Value: "ap-south-2", Label: "Asia Pacific (Hyderabad)"},
	{Value: "ap-southeast-1", Label: "Asia Pacific (Singapore)"},
	{Value: "ap-southeast-2", Label: "Asia Pacific (Sydney)"},
	{Value: "ap-southeast-3", Label: "Asia Pacific (Jakarta)"},
	{Value: "ap-southeast-4", Label: "Asia Pacific (Melbourne)"},
	{Value: "ap-southeast-5", Label: "Asia Pacific (Malaysia)"},
	{Value: "ap-southeast-6", Label: "Asia Pacific (New Zealand)"},
	{Value: "ap-southeast-7", Label: "Asia Pacific (Thailand)"},
	{Value: "ca-central-1", Label: "Canada (Central)"},
	{Value: "ca-west-1", Label: "Canada West (Calgary)"},
	{Value: "cn-north-1", Label: "China (Beijing)"},
	{Value: "cn-northwest-1", Label: "China (Ningxia)"},
	{Value: "eu-central-1", Label: "Europe (Frankfurt)"},
	{Value: "eu-central-2", Label: "Europe (Zurich)"},
	{Value: "eu-north-1", Label: "Europe (Stockholm)"},
	{Value: "eu-south-1", Label: "Europe (Milan)"},
	{Value: "eu-south-2", Label: "Europe (Spain)"},
	{Value: "eu-west-1", Label: "Europe (Ireland)"},
	{Value: "eu-west-2", Label: "Europe (London)"},
	{Value: "eu-west-3", Label: "Europe (Paris)"},
	{Value: "eusc-de-east-1", Label: "European Sovereign Cloud (Brandenburg)"},
	{Value: "il-central-1", Label: "Israel (Tel Aviv)"},
	{Value: "me-central-1", Label: "Middle East (UAE)"},
	{Value: "me-south-1", Label: "Middle East (Bahrain)"},
	{Value: "mx-central-1", Label: "Mexico (Central)"},
	{Value: "sa-east-1", Label: "South America (São Paulo)"},
	{Value: "us-east-1", Label: "US East (N. Virginia)"},
	{Value: "us-east-2", Label: "US East (Ohio)"},
	{Value: "us-gov-east-1", Label: "AWS GovCloud (US-East)"},
	{Value: "us-gov-west-1", Label: "AWS GovCloud (US-West)"},
	{Value: "us-west-1", Label: "US West (N. California)"},
	{Value: "us-west-2", Label: "US West (Oregon)"},
}

func RegionLabel(region string) string {
	for _, item := range regionCatalog {
		if item.Value == region {
			return item.Label
		}
	}
	return region
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

func (s *RegionService) GetRegionCatalog() []RegionInfo {
	return append([]RegionInfo(nil), regionCatalog...)
}

func (s *RegionService) fallbackRegions() []RegionInfo {
	return s.GetRegionCatalog()
}

var _ = time.Now
