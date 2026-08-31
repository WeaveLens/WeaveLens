package client

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Clients struct {
	EC2   EC2API
	RDS   RDSAPI
	ELBv2 ELBv2API
	STS   STSAPI
}

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) BuildClients(cfg aws.Config) *Clients {
	return &Clients{
		EC2:   ec2.NewFromConfig(cfg),
		RDS:   rds.NewFromConfig(cfg),
		ELBv2: elasticloadbalancingv2.NewFromConfig(cfg),
		STS:   sts.NewFromConfig(cfg),
	}
}
