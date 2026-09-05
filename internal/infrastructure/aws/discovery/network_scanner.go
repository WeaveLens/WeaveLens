package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elip/WeaveLens/internal/domain/resource"
	awsclient "github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("Network", func(c *awsclient.Clients, region string) Scanner { return NewNetworkScanner(c.EC2, region) })
}

type NetworkScanner struct {
	client EC2API
	region string
}

func NewNetworkScanner(client EC2API, region string) *NetworkScanner {
	return &NetworkScanner{client: client, region: region}
}

func (s *NetworkScanner) Name() string {
	return "Network"
}

func (s *NetworkScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	var resources []*resource.Resource

	vpcs, err := s.scanVPCs(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "VPC", Err: err}
	}
	resources = append(resources, vpcs...)

	subnets, err := s.scanSubnets(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "Subnet", Err: err}
	}
	resources = append(resources, subnets...)

	routeTables, err := s.scanRouteTables(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "RouteTable", Err: err}
	}
	resources = append(resources, routeTables...)

	igws, err := s.scanInternetGateways(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "InternetGateway", Err: err}
	}
	resources = append(resources, igws...)

	natGateways, err := s.scanNATGateways(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "NATGateway", Err: err}
	}
	resources = append(resources, natGateways...)

	securityGroups, err := s.scanSecurityGroups(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "SecurityGroup", Err: err}
	}
	resources = append(resources, securityGroups...)

	return resources, nil
}

func (s *NetworkScanner) scanVPCs(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeVpcsPaginator(s.client, &ec2.DescribeVpcsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, vpc := range page.Vpcs {
			if vpc.VpcId == nil {
				continue
			}
			name, tags := extractTags(vpc.Tags)
			if name == "" {
				name = *vpc.VpcId
			}

			res, err := resource.NewResource(
				resource.ResourceID(*vpc.VpcId),
				resource.ResourceType("VPC"),
				resource.CategoryNetwork,
				name,
				resource.WithMetadata(map[string]string{
					"cidr": safePtr(vpc.CidrBlock),
				}),
				resource.WithTags(tags),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func (s *NetworkScanner) scanSubnets(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeSubnetsPaginator(s.client, &ec2.DescribeSubnetsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, subnet := range page.Subnets {
			if subnet.SubnetId == nil {
				continue
			}
			name, tags := extractTags(subnet.Tags)
			if name == "" {
				name = *subnet.SubnetId
			}

			metadata := map[string]string{
				"vpc_id":     safePtr(subnet.VpcId),
				"cidr_block": safePtr(subnet.CidrBlock),
				"az":         safePtr(subnet.AvailabilityZone),
			}

			res, err := resource.NewResource(
				resource.ResourceID(*subnet.SubnetId),
				resource.ResourceType("Subnet"),
				resource.CategoryNetwork,
				name,
				resource.WithMetadata(metadata),
				resource.WithTags(tags),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func (s *NetworkScanner) scanRouteTables(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeRouteTablesPaginator(s.client, &ec2.DescribeRouteTablesInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, rt := range page.RouteTables {
			if rt.RouteTableId == nil {
				continue
			}
			name, tags := extractTags(rt.Tags)
			if name == "" {
				name = *rt.RouteTableId
			}
			metadata := map[string]string{"vpc_id": safePtr(rt.VpcId)}
			var subnetIDs []string
			var routeTargetIDs []string
			for _, association := range rt.Associations {
				if association.SubnetId != nil {
					subnetIDs = append(subnetIDs, *association.SubnetId)
				}
			}
			for _, route := range rt.Routes {
				for _, target := range []string{
					safePtr(route.GatewayId), safePtr(route.NatGatewayId),
					safePtr(route.TransitGatewayId), safePtr(route.VpcPeeringConnectionId),
					safePtr(route.NetworkInterfaceId),
				} {
					if target != "" {
						routeTargetIDs = append(routeTargetIDs, target)
					}
				}
			}
			if len(subnetIDs) > 0 {
				metadata["subnet_ids"] = strings.Join(subnetIDs, ",")
			}
			if len(routeTargetIDs) > 0 {
				metadata["route_target_ids"] = strings.Join(routeTargetIDs, ",")
			}

			res, err := resource.NewResource(
				resource.ResourceID(*rt.RouteTableId),
				resource.ResourceType("RouteTable"),
				resource.CategoryNetwork,
				name,
				resource.WithMetadata(metadata),
				resource.WithTags(tags),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func (s *NetworkScanner) scanInternetGateways(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeInternetGatewaysPaginator(s.client, &ec2.DescribeInternetGatewaysInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, igw := range page.InternetGateways {
			if igw.InternetGatewayId == nil {
				continue
			}
			name, tags := extractTags(igw.Tags)
			if name == "" {
				name = *igw.InternetGatewayId
			}

			var vpcID string
			for _, att := range igw.Attachments {
				if att.VpcId != nil {
					vpcID = *att.VpcId
					break
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(*igw.InternetGatewayId),
				resource.ResourceType("InternetGateway"),
				resource.CategoryNetwork,
				name,
				resource.WithMetadata(map[string]string{
					"vpc_id": vpcID,
				}),
				resource.WithTags(tags),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func (s *NetworkScanner) scanNATGateways(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeNatGatewaysPaginator(s.client, &ec2.DescribeNatGatewaysInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, nat := range page.NatGateways {
			if nat.NatGatewayId == nil {
				continue
			}
			name, tags := extractTags(nat.Tags)
			if name == "" {
				name = *nat.NatGatewayId
			}

			metadata := map[string]string{
				"state": string(nat.State),
			}
			if nat.VpcId != nil {
				metadata["vpc_id"] = *nat.VpcId
			}
			if nat.SubnetId != nil {
				metadata["subnet_id"] = *nat.SubnetId
			}

			res, err := resource.NewResource(
				resource.ResourceID(*nat.NatGatewayId),
				resource.ResourceType("NATGateway"),
				resource.CategoryNetwork,
				name,
				resource.WithMetadata(metadata),
				resource.WithTags(tags),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func (s *NetworkScanner) scanSecurityGroups(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeSecurityGroupsPaginator(s.client, &ec2.DescribeSecurityGroupsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, sg := range page.SecurityGroups {
			if sg.GroupId == nil || sg.GroupName == nil {
				continue
			}
			name, tags := extractTags(sg.Tags)
			if name == "" {
				name = *sg.GroupName
			}
			metadata := map[string]string{
				"group_id":   *sg.GroupId,
				"group_name": *sg.GroupName,
				"vpc_id":     safePtr(sg.VpcId),
			}
			var referencedGroupIDs []string
			for _, permission := range append(sg.IpPermissions, sg.IpPermissionsEgress...) {
				for _, pair := range permission.UserIdGroupPairs {
					if pair.GroupId != nil {
						referencedGroupIDs = append(referencedGroupIDs, *pair.GroupId)
					}
				}
			}
			if len(referencedGroupIDs) > 0 {
				metadata["referenced_group_ids"] = strings.Join(referencedGroupIDs, ",")
			}

			res, err := resource.NewResource(
				resource.ResourceID(*sg.GroupId),
				resource.ResourceType("SecurityGroup"),
				resource.CategorySecurity,
				name,
				resource.WithMetadata(metadata),
				resource.WithTags(tags),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func extractTags(tags []ec2types.Tag) (string, map[string]string) {
	result := make(map[string]string)
	name := ""
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
			if *tag.Key == "Name" {
				name = *tag.Value
			}
		}
	}
	return name, result
}

func safePtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func resourceIDFromARN(arn string) string {
	if idx := strings.LastIndex(arn, ":"); idx >= 0 && idx < len(arn)-1 {
		return arn[idx+1:]
	}
	return arn
}

type EC2API interface {
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
	DescribeInternetGateways(ctx context.Context, params *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error)
	DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
}

var _ = ec2types.Vpc{}
