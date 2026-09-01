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
	if cfg.Region == "" {
		return nil, nil
	}

	client := ec2.NewFromConfig(cfg)
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
	strategyRegion := cfg.Region
	if strategyRegion == "" {
		regions, err := availableRegions(context.Background(), cfg.AWSConfig)
		if err == nil && len(regions) > 0 {
			var scanners []Scanner
			for _, region := range regions {
				regionalCfg := cfg.AWSConfig
				regionalCfg.Region = region
				regionalClients := client.NewFactory().BuildClients(regionalCfg)
				scanners = append(scanners, buildRegionScanners(regionalClients, region)...)
			}
			return NewService(scanners, NewRelationshipBuilder())
		}
		strategyRegion = "us-east-1"
	}

	networkScanner := NewNetworkScanner(cfg.Clients.EC2, strategyRegion)
	computeScanner := NewComputeScanner(cfg.Clients.EC2, strategyRegion)
	databaseScanner := NewDatabaseScanner(cfg.Clients.RDS, strategyRegion)
	loadBalancerScanner := NewLoadBalancerScanner(cfg.Clients.ELBv2, strategyRegion)
	relationshipBuilder := NewRelationshipBuilder()

	return NewService(
		[]Scanner{networkScanner, computeScanner, databaseScanner, loadBalancerScanner},
		relationshipBuilder,
	)
}

func NewServiceFromConfigWithResilience(cfg ServiceConfigInput, resilienceCfg ServiceConfig) *Service {
	strategyRegion := cfg.Region
	if strategyRegion == "" {
		regions, err := availableRegions(context.Background(), cfg.AWSConfig)
		if err == nil && len(regions) > 0 {
			var scanners []Scanner
			for _, region := range regions {
				regionalCfg := cfg.AWSConfig
				regionalCfg.Region = region
				regionalClients := client.NewFactory().BuildClients(regionalCfg)
				scanners = append(scanners, buildRegionScanners(regionalClients, region)...)
			}
			return NewServiceWithConfig(scanners, NewRelationshipBuilder(), resilienceCfg)
		}
		strategyRegion = "us-east-1"
	}

	networkScanner := NewNetworkScanner(cfg.Clients.EC2, strategyRegion)
	computeScanner := NewComputeScanner(cfg.Clients.EC2, strategyRegion)
	databaseScanner := NewDatabaseScanner(cfg.Clients.RDS, strategyRegion)
	loadBalancerScanner := NewLoadBalancerScanner(cfg.Clients.ELBv2, strategyRegion)
	relationshipBuilder := NewRelationshipBuilder()

	return NewServiceWithConfig(
		[]Scanner{networkScanner, computeScanner, databaseScanner, loadBalancerScanner},
		relationshipBuilder,
		resilienceCfg,
	)
}
