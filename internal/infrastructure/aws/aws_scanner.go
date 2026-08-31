package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

type AWSScanner struct {
	clients *client.Clients
}

func NewAWSScanner(clients *client.Clients) *AWSScanner {
	return &AWSScanner{clients: clients}
}

func (s *AWSScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	var all []*resource.Resource

	vpcs, err := s.scanVPCs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan VPCs: %w", err)
	}
	all = append(all, vpcs...)

	subnets, err := s.scanSubnets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan subnets: %w", err)
	}
	all = append(all, subnets...)

	routeTables, err := s.scanRouteTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan route tables: %w", err)
	}
	all = append(all, routeTables...)

	igws, err := s.scanInternetGateways(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan internet gateways: %w", err)
	}
	all = append(all, igws...)

	natGateways, err := s.scanNATGateways(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan NAT gateways: %w", err)
	}
	all = append(all, natGateways...)

	sgs, err := s.scanSecurityGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan security groups: %w", err)
	}
	all = append(all, sgs...)

	ec2s, err := s.scanEC2Instances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan EC2 instances: %w", err)
	}
	all = append(all, ec2s...)

	rdss, err := s.scanRDSInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan RDS instances: %w", err)
	}
	all = append(all, rdss...)

	albs, err := s.scanALBs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan ALBs: %w", err)
	}
	all = append(all, albs...)

	return all, nil
}

func (s *AWSScanner) scanVPCs(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.EC2.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	return mapVPCs(resp.Vpcs), nil
}

func (s *AWSScanner) scanSubnets(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.EC2.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, err
	}
	return mapSubnets(resp.Subnets), nil
}

func (s *AWSScanner) scanRouteTables(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.EC2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{})
	if err != nil {
		return nil, err
	}
	return mapRouteTables(resp.RouteTables), nil
}

func (s *AWSScanner) scanInternetGateways(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.EC2.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{})
	if err != nil {
		return nil, err
	}
	return mapInternetGateways(resp.InternetGateways), nil
}

func (s *AWSScanner) scanNATGateways(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.EC2.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return nil, err
	}
	return mapNATGateways(resp.NatGateways), nil
}

func (s *AWSScanner) scanSecurityGroups(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.EC2.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, err
	}
	return mapSecurityGroups(resp.SecurityGroups), nil
}

func (s *AWSScanner) scanEC2Instances(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}
	return mapEC2Instances(resp.Reservations), nil
}

func (s *AWSScanner) scanRDSInstances(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.RDS.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, err
	}
	return mapRDSInstances(resp.DBInstances), nil
}

func (s *AWSScanner) scanALBs(ctx context.Context) ([]*resource.Resource, error) {
	resp, err := s.clients.ELBv2.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}
	return mapALBs(resp.LoadBalancers), nil
}
