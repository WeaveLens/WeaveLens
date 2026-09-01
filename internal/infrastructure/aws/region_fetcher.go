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
			Label: formatRegionLabel(*r.RegionName),
		})
	}

	sort.Slice(regions, func(i, j int) bool {
		return regions[i].Value < regions[j].Value
	})

	return regions, nil
}

func formatRegionLabel(region string) string {
	labels := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-east-2":      "US East (Ohio)",
		"us-west-1":      "US West (N. California)",
		"us-west-2":      "US West (Oregon)",
		"us-gov-east-1":  "AWS GovCloud (US-East)",
		"us-gov-west-1":  "AWS GovCloud (US-West)",
		"ca-central-1":   "Canada (Central)",
		"ca-west-1":      "Canada West (Calgary)",
		"sa-east-1":      "South America (São Paulo)",
		"eu-west-1":      "Europe (Ireland)",
		"eu-west-2":      "Europe (London)",
		"eu-west-3":      "Europe (Paris)",
		"eu-central-1":   "Europe (Frankfurt)",
		"eu-central-2":   "Europe (Zurich)",
		"eu-south-1":     "Europe (Milan)",
		"eu-south-2":     "Europe (Spain)",
		"eu-north-1":     "Europe (Stockholm)",
		"me-south-1":     "Middle East (Bahrain)",
		"me-central-1":   "Middle East (UAE)",
		"il-central-1":   "Israel (Tel Aviv)",
		"af-south-1":     "Africa (Cape Town)",
		"ap-east-1":      "Asia Pacific (Hong Kong)",
		"ap-east-2":      "Asia Pacific (Taipei)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ap-south-2":     "Asia Pacific (Hyderabad)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-southeast-3": "Asia Pacific (Jakarta)",
		"ap-southeast-4": "Asia Pacific (Melbourne)",
		"ap-southeast-5": "Asia Pacific (Malaysia)",
		"ap-southeast-6": "Asia Pacific (New Zealand)",
		"ap-southeast-7": "Asia Pacific (Thailand)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-northeast-3": "Asia Pacific (Osaka)",
		"cn-north-1":     "China (Beijing)",
		"cn-northwest-1": "China (Ningxia)",
		"mx-central-1":   "Mexico (Central)",
	}
	if label, ok := labels[region]; ok {
		return label
	}
	return region
}
