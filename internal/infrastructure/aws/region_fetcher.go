package aws

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elip/WeaveLens/internal/application/service"
)

type RegionFetcher struct {
	client *ec2.Client
}

func NewRegionFetcher(client *ec2.Client) *RegionFetcher {
	return &RegionFetcher{client: client}
}

func (f *RegionFetcher) FetchRegions(ctx context.Context) ([]service.RegionInfo, error) {
	output, err := f.client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	var regions []service.RegionInfo
	for _, r := range output.Regions {
		if r.RegionName == nil {
			continue
		}
		state := ""
		if r.OptInStatus != nil {
			state = *r.OptInStatus
		}
		if state == "not-opted-in" {
			continue
		}
		regions = append(regions, service.RegionInfo{
			Value: *r.RegionName,
			Label: service.RegionLabel(*r.RegionName),
		})
	}

	sort.Slice(regions, func(i, j int) bool {
		return regions[i].Value < regions[j].Value
	})

	return regions, nil
}
