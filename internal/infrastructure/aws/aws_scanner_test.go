package aws

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

func strPtr(s string) *string {
	return &s
}

func TestMapVPCs(t *testing.T) {
	vpcs := []ec2types.Vpc{
		{VpcId: strPtr("vpc-1"), CidrBlock: strPtr("10.0.0.0/16")},
		{VpcId: strPtr("vpc-2"), CidrBlock: strPtr("10.1.0.0/16")},
	}

	result := mapVPCs(vpcs)

	if len(result) != 2 {
		t.Errorf("mapVPCs() length = %v, want 2", len(result))
	}

	if result[0].ID() != "vpc-1" {
		t.Errorf("mapVPCs()[0].ID() = %v, want vpc-1", result[0].ID())
	}
	if result[0].Type() != "VPC" {
		t.Errorf("mapVPCs()[0].Type() = %v, want VPC", result[0].Type())
	}
	if result[0].Category() != resource.CategoryNetwork {
		t.Errorf("mapVPCs()[0].Category() = %v, want network", result[0].Category())
	}
}

func TestMapSubnets(t *testing.T) {
	subnets := []ec2types.Subnet{
		{SubnetId: strPtr("subnet-1"), VpcId: strPtr("vpc-1"), CidrBlock: strPtr("10.0.1.0/24")},
	}

	result := mapSubnets(subnets)

	if len(result) != 1 {
		t.Errorf("mapSubnets() length = %v, want 1", len(result))
	}

	if result[0].ID() != "subnet-1" {
		t.Errorf("mapSubnets()[0].ID() = %v, want subnet-1", result[0].ID())
	}
}

func TestMapRouteTables(t *testing.T) {
	rts := []ec2types.RouteTable{
		{RouteTableId: strPtr("rt-1"), VpcId: strPtr("vpc-1")},
	}

	result := mapRouteTables(rts)

	if len(result) != 1 {
		t.Errorf("mapRouteTables() length = %v, want 1", len(result))
	}
}

func TestMapInternetGateways(t *testing.T) {
	igws := []ec2types.InternetGateway{
		{InternetGatewayId: strPtr("igw-1")},
	}

	result := mapInternetGateways(igws)

	if len(result) != 1 {
		t.Errorf("mapInternetGateways() length = %v, want 1", len(result))
	}
}

func TestMapNATGateways(t *testing.T) {
	nats := []ec2types.NatGateway{
		{NatGatewayId: strPtr("nat-1"), State: ec2types.NatGatewayStateAvailable},
	}

	result := mapNATGateways(nats)

	if len(result) != 1 {
		t.Errorf("mapNATGateways() length = %v, want 1", len(result))
	}
}

func TestMapSecurityGroups(t *testing.T) {
	sgs := []ec2types.SecurityGroup{
		{GroupId: strPtr("sg-1"), GroupName: strPtr("my-sg"), VpcId: strPtr("vpc-1")},
	}

	result := mapSecurityGroups(sgs)

	if len(result) != 1 {
		t.Errorf("mapSecurityGroups() length = %v, want 1", len(result))
	}
	if result[0].Category() != resource.CategorySecurity {
		t.Errorf("mapSecurityGroups()[0].Category() = %v, want security", result[0].Category())
	}
}

func TestMapEC2Instances(t *testing.T) {
	reservations := []ec2types.Reservation{
		{
			Instances: []ec2types.Instance{
				{
					InstanceId:     strPtr("i-1"),
					ImageId:        strPtr("ami-123"),
					InstanceType:   ec2types.InstanceTypeT2Micro,
					State: &ec2types.InstanceState{
						Name: ec2types.InstanceStateNameRunning,
					},
					Tags: []ec2types.Tag{
						{Key: strPtr("Name"), Value: strPtr("MyInstance")},
					},
				},
			},
		},
	}

	result := mapEC2Instances(reservations)

	if len(result) != 1 {
		t.Errorf("mapEC2Instances() length = %v, want 1", len(result))
	}
	if result[0].Type() != "EC2" {
		t.Errorf("mapEC2Instances()[0].Type() = %v, want EC2", result[0].Type())
	}
	if result[0].Category() != resource.CategoryCompute {
		t.Errorf("mapEC2Instances()[0].Category() = %v, want compute", result[0].Category())
	}
	if result[0].Name() != "MyInstance" {
		t.Errorf("mapEC2Instances()[0].Name() = %v, want MyInstance", result[0].Name())
	}
}

func TestMapRDSInstances(t *testing.T) {
	dbInstances := []rdstypes.DBInstance{
		{
			DBInstanceIdentifier: strPtr("db-1"),
			Engine:               strPtr("mysql"),
			DBInstanceClass:      strPtr("db.t3.micro"),
			DBInstanceStatus:     strPtr("available"),
		},
	}

	result := mapRDSInstances(dbInstances)

	if len(result) != 1 {
		t.Errorf("mapRDSInstances() length = %v, want 1", len(result))
	}
	if result[0].Type() != "RDS" {
		t.Errorf("mapRDSInstances()[0].Type() = %v, want RDS", result[0].Type())
	}
	if result[0].Category() != resource.CategoryDatabase {
		t.Errorf("mapRDSInstances()[0].Category() = %v, want database", result[0].Category())
	}
}

func TestMapALBs(t *testing.T) {
	lbs := []elbv2types.LoadBalancer{
		{
			LoadBalancerArn:  strPtr("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb"),
			LoadBalancerName: strPtr("my-alb"),
			DNSName:          strPtr("my-alb-123.elb.us-east-1.amazonaws.com"),
			Type:             elbv2types.LoadBalancerTypeEnumApplication,
			Scheme:           elbv2types.LoadBalancerSchemeEnumInternetFacing,
		},
	}

	result := mapALBs(lbs)

	if len(result) != 1 {
		t.Errorf("mapALBs() length = %v, want 1", len(result))
	}
	if result[0].Type() != "ALB" {
		t.Errorf("mapALBs()[0].Type() = %v, want ALB", result[0].Type())
	}
	if result[0].Category() != resource.CategoryNetwork {
		t.Errorf("mapALBs()[0].Category() = %v, want network", result[0].Category())
	}
}
