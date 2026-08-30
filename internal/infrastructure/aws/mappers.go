package aws

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

func mapVPCs(vpcs []ec2types.Vpc) []*resource.Resource {
	var result []*resource.Resource
	for _, vpc := range vpcs {
		if vpc.VpcId == nil || vpc.CidrBlock == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*vpc.VpcId),
			resource.ResourceType("VPC"),
			resource.CategoryNetwork,
			*vpc.VpcId,
			resource.WithMetadata(map[string]string{
				"cidr": *vpc.CidrBlock,
			}),
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}

func mapSubnets(subnets []ec2types.Subnet) []*resource.Resource {
	var result []*resource.Resource
	for _, subnet := range subnets {
		if subnet.SubnetId == nil || subnet.VpcId == nil || subnet.CidrBlock == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*subnet.SubnetId),
			resource.ResourceType("Subnet"),
			resource.CategoryNetwork,
			*subnet.SubnetId,
			resource.WithMetadata(map[string]string{
				"vpc_id":     *subnet.VpcId,
				"cidr_block": *subnet.CidrBlock,
			}),
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}

func mapRouteTables(routeTables []ec2types.RouteTable) []*resource.Resource {
	var result []*resource.Resource
	for _, rt := range routeTables {
		if rt.RouteTableId == nil || rt.VpcId == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*rt.RouteTableId),
			resource.ResourceType("RouteTable"),
			resource.CategoryNetwork,
			*rt.RouteTableId,
			resource.WithMetadata(map[string]string{
				"vpc_id": *rt.VpcId,
			}),
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}

func mapInternetGateways(igws []ec2types.InternetGateway) []*resource.Resource {
	var result []*resource.Resource
	for _, igw := range igws {
		if igw.InternetGatewayId == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*igw.InternetGatewayId),
			resource.ResourceType("InternetGateway"),
			resource.CategoryNetwork,
			*igw.InternetGatewayId,
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}

func mapNATGateways(nats []ec2types.NatGateway) []*resource.Resource {
	var result []*resource.Resource
	for _, nat := range nats {
		if nat.NatGatewayId == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*nat.NatGatewayId),
			resource.ResourceType("NATGateway"),
			resource.CategoryNetwork,
			*nat.NatGatewayId,
			resource.WithMetadata(map[string]string{
				"state": string(nat.State),
			}),
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}

func mapSecurityGroups(sgs []ec2types.SecurityGroup) []*resource.Resource {
	var result []*resource.Resource
	for _, sg := range sgs {
		if sg.GroupId == nil || sg.GroupName == nil || sg.VpcId == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*sg.GroupId),
			resource.ResourceType("SecurityGroup"),
			resource.CategorySecurity,
			*sg.GroupName,
			resource.WithMetadata(map[string]string{
				"group_id":   *sg.GroupId,
				"vpc_id":     *sg.VpcId,
				"group_name": *sg.GroupName,
			}),
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}

func mapEC2Instances(reservations []ec2types.Reservation) []*resource.Resource {
	var result []*resource.Resource
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceId == nil || instance.ImageId == nil {
				continue
			}

			name := ""
			for _, tag := range instance.Tags {
				if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
					name = *tag.Value
					break
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(*instance.InstanceId),
				resource.ResourceType("EC2"),
				resource.CategoryCompute,
				name,
				resource.WithMetadata(map[string]string{
					"instance_type": string(instance.InstanceType),
					"state":         string(instance.State.Name),
					"image_id":      *instance.ImageId,
				}),
			)
			if err != nil {
				continue
			}
			result = append(result, res)
		}
	}
	return result
}

func mapRDSInstances(dbInstances []rdstypes.DBInstance) []*resource.Resource {
	var result []*resource.Resource
	for _, db := range dbInstances {
		if db.DBInstanceIdentifier == nil || db.Engine == nil || db.DBInstanceClass == nil || db.DBInstanceStatus == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*db.DBInstanceIdentifier),
			resource.ResourceType("RDS"),
			resource.CategoryDatabase,
			*db.DBInstanceIdentifier,
			resource.WithMetadata(map[string]string{
				"engine":        *db.Engine,
				"instance_class": *db.DBInstanceClass,
				"status":        *db.DBInstanceStatus,
			}),
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}

func mapALBs(lbs []elbv2types.LoadBalancer) []*resource.Resource {
	var result []*resource.Resource
	for _, lb := range lbs {
		if lb.LoadBalancerArn == nil || lb.LoadBalancerName == nil || lb.DNSName == nil {
			continue
		}
		res, err := resource.NewResource(
			resource.ResourceID(*lb.LoadBalancerArn),
			resource.ResourceType("ALB"),
			resource.CategoryNetwork,
			*lb.LoadBalancerName,
			resource.WithMetadata(map[string]string{
				"dns_name":    *lb.DNSName,
				"type":        string(lb.Type),
				"scheme":      string(lb.Scheme),
			}),
		)
		if err != nil {
			continue
		}
		result = append(result, res)
	}
	return result
}
