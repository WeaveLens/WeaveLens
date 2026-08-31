package discovery

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type mockEC2Client struct {
	describeVpcsOutput    *ec2.DescribeVpcsOutput
	describeVpcsErr       error
	describeSubnetsOutput *ec2.DescribeSubnetsOutput
	describeSubnetsErr    error
}

func (m *mockEC2Client) DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	return m.describeVpcsOutput, m.describeVpcsErr
}

func (m *mockEC2Client) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	return m.describeSubnetsOutput, m.describeSubnetsErr
}

func (m *mockEC2Client) DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
	return &ec2.DescribeRouteTablesOutput{}, nil
}

func (m *mockEC2Client) DescribeInternetGateways(ctx context.Context, params *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
	return &ec2.DescribeInternetGatewaysOutput{}, nil
}

func (m *mockEC2Client) DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error) {
	return &ec2.DescribeNatGatewaysOutput{}, nil
}

func (m *mockEC2Client) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{}, nil
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{}, nil
}

func TestNetworkScanner_ScanVPCs(t *testing.T) {
	mockClient := &mockEC2Client{
		describeVpcsOutput: &ec2.DescribeVpcsOutput{
			Vpcs: []ec2types.Vpc{
				{
					VpcId:     aws.String("vpc-123"),
					CidrBlock: aws.String("10.0.0.0/16"),
					Tags: []ec2types.Tag{
						{Key: aws.String("Name"), Value: aws.String("test-vpc")},
					},
				},
			},
		},
		describeSubnetsOutput: &ec2.DescribeSubnetsOutput{},
	}

	scanner := NewNetworkScanner(mockClient)
	resources, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}

	found := false
	for _, res := range resources {
		if res.ID() == "vpc-123" {
			found = true
			if res.Type() != "VPC" {
				t.Errorf("Type() = %v, want VPC", res.Type())
			}
			if res.Name() != "test-vpc" {
				t.Errorf("Name() = %v, want test-vpc", res.Name())
			}
			if res.Category() != resource.CategoryNetwork {
				t.Errorf("Category() = %v, want network", res.Category())
			}
		}
	}
	if !found {
		t.Error("Expected to find vpc-123 in scan results")
	}
}

func TestNetworkScanner_ScanEmpty(t *testing.T) {
	mockClient := &mockEC2Client{
		describeVpcsOutput:    &ec2.DescribeVpcsOutput{Vpcs: []ec2types.Vpc{}},
		describeSubnetsOutput: &ec2.DescribeSubnetsOutput{},
	}

	scanner := NewNetworkScanner(mockClient)
	resources, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("Expected 0 resources, got %d", len(resources))
	}
}

func TestNetworkScanner_APIError(t *testing.T) {
	mockClient := &mockEC2Client{
		describeVpcsErr: errors.New("AccessDenied: not authorized"),
	}

	scanner := NewNetworkScanner(mockClient)
	_, err := scanner.Scan(context.Background())
	if err == nil {
		t.Fatal("Scan() expected error, got nil")
	}

	var scannerErr *ScannerError
	if !errors.As(err, &scannerErr) {
		t.Errorf("Expected ScannerError, got %T", err)
	}
}

func TestNetworkScanner_ContextCancellation(t *testing.T) {
	mockClient := &mockEC2Client{
		describeVpcsOutput:    &ec2.DescribeVpcsOutput{Vpcs: []ec2types.Vpc{}},
		describeSubnetsOutput: &ec2.DescribeSubnetsOutput{},
	}

	scanner := NewNetworkScanner(mockClient)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := scanner.Scan(ctx)
	if err == nil {
		t.Fatal("Scan() expected error for cancelled context")
	}
}
