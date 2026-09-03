package discovery

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

type ServiceConfigInput struct {
	Clients   *client.Clients
	Region    string
	AWSConfig aws.Config
}

func buildRegionScanners(clients *client.Clients, region string) []Scanner {
	return []Scanner{
		NewNetworkScanner(clients.EC2, region),
		NewComputeScanner(clients.EC2, region),
		NewDatabaseScanner(clients.RDS, region),
		NewLoadBalancerScanner(clients.ELBv2, region),
	}
}

func availableRegions(ctx context.Context, cfg aws.Config) ([]string, error) {
	describeCfg := cfg
	if describeCfg.Region == "" {
		describeCfg.Region = "us-east-1"
	}

	client := ec2.NewFromConfig(describeCfg)
	output, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{AllRegions: aws.Bool(true)})
	if err != nil {
		return nil, err
	}

	regions := make([]string, 0, len(output.Regions))
	for _, region := range output.Regions {
		if region.RegionName != nil && *region.RegionName != "" {
			regions = append(regions, *region.RegionName)
		}
	}
	return regions, nil
}

func NewServiceFromConfig(cfg ServiceConfigInput) *Service {
	return NewServiceFromConfigWithResilience(cfg, DefaultServiceConfig())
}

func NewServiceFromConfigWithResilience(cfg ServiceConfigInput, resilienceCfg ServiceConfig) *Service {
	factory := func(region string) ([]Scanner, error) {
		regionalCfg := cfg.AWSConfig
		regionalCfg.Region = region
		regionalClients := client.NewFactory().BuildClients(regionalCfg)
		return buildRegionScanners(regionalClients, region), nil
	}

	return NewServiceDynamic(cfg.AWSConfig, factory, NewRelationshipBuilder(), resilienceCfg)
}
