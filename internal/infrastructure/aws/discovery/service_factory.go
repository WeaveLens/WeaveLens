package discovery

import (
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

type ServiceConfigInput struct {
	Clients *client.Clients
	Region  string
}

func NewServiceFromConfig(cfg ServiceConfigInput) *Service {
	networkScanner := NewNetworkScanner(cfg.Clients.EC2, cfg.Region)
	computeScanner := NewComputeScanner(cfg.Clients.EC2, cfg.Region)
	databaseScanner := NewDatabaseScanner(cfg.Clients.RDS, cfg.Region)
	loadBalancerScanner := NewLoadBalancerScanner(cfg.Clients.ELBv2, cfg.Region)
	relationshipBuilder := NewRelationshipBuilder()

	return NewService(
		[]Scanner{networkScanner, computeScanner, databaseScanner, loadBalancerScanner},
		relationshipBuilder,
	)
}

func NewServiceFromConfigWithResilience(cfg ServiceConfigInput, resilienceCfg ServiceConfig) *Service {
	networkScanner := NewNetworkScanner(cfg.Clients.EC2, cfg.Region)
	computeScanner := NewComputeScanner(cfg.Clients.EC2, cfg.Region)
	databaseScanner := NewDatabaseScanner(cfg.Clients.RDS, cfg.Region)
	loadBalancerScanner := NewLoadBalancerScanner(cfg.Clients.ELBv2, cfg.Region)
	relationshipBuilder := NewRelationshipBuilder()

	return NewServiceWithConfig(
		[]Scanner{networkScanner, computeScanner, databaseScanner, loadBalancerScanner},
		relationshipBuilder,
		resilienceCfg,
	)
}
