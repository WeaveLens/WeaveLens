package discovery

import (
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

type ServiceConfig struct {
	Clients *client.Clients
}

func NewServiceFromConfig(cfg ServiceConfig) *Service {
	networkScanner := NewNetworkScanner(cfg.Clients.EC2)
	computeScanner := NewComputeScanner(cfg.Clients.EC2)
	databaseScanner := NewDatabaseScanner(cfg.Clients.RDS)
	loadBalancerScanner := NewLoadBalancerScanner(cfg.Clients.ELBv2)
	relationshipBuilder := NewRelationshipBuilder()

	return NewService(
		[]Scanner{networkScanner, computeScanner, databaseScanner, loadBalancerScanner},
		relationshipBuilder,
	)
}
