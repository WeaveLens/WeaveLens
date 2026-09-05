package discovery

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type mockComputeClient struct{}

func (mockComputeClient) DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{Reservations: []ec2types.Reservation{{Instances: []ec2types.Instance{{
		InstanceId: aws.String("i-123"),
		NetworkInterfaces: []ec2types.InstanceNetworkInterface{
			{NetworkInterfaceId: aws.String("eni-123")},
			{NetworkInterfaceId: aws.String("eni-456")},
		},
	}}}}}, nil
}

type mockAttachmentClient struct{}

func (mockAttachmentClient) DescribeVpcEndpoints(context.Context, *ec2.DescribeVpcEndpointsInput, ...func(*ec2.Options)) (*ec2.DescribeVpcEndpointsOutput, error) {
	return &ec2.DescribeVpcEndpointsOutput{}, nil
}

func (mockAttachmentClient) DescribeTransitGatewayAttachments(context.Context, *ec2.DescribeTransitGatewayAttachmentsInput, ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayAttachmentsOutput, error) {
	return &ec2.DescribeTransitGatewayAttachmentsOutput{}, nil
}

func (mockAttachmentClient) DescribeVpcPeeringConnections(context.Context, *ec2.DescribeVpcPeeringConnectionsInput, ...func(*ec2.Options)) (*ec2.DescribeVpcPeeringConnectionsOutput, error) {
	return &ec2.DescribeVpcPeeringConnectionsOutput{}, nil
}

func (mockAttachmentClient) DescribeNetworkInterfaces(context.Context, *ec2.DescribeNetworkInterfacesInput, ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return &ec2.DescribeNetworkInterfacesOutput{NetworkInterfaces: []ec2types.NetworkInterface{{
		NetworkInterfaceId: aws.String("eni-123"),
		Attachment:         &ec2types.NetworkInterfaceAttachment{InstanceId: aws.String("i-123")},
	}}}, nil
}

func (mockAttachmentClient) DescribeVolumes(context.Context, *ec2.DescribeVolumesInput, ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	return &ec2.DescribeVolumesOutput{}, nil
}

func TestEC2AndENIScanners_EmitRelationshipMetadata(t *testing.T) {
	instances, err := NewComputeScanner(mockComputeClient{}, "us-east-1").Scan(context.Background())
	if err != nil {
		t.Fatalf("Compute Scan() unexpected error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("Compute Scan() returned %d resources, want 1", len(instances))
	}
	if got := instances[0].Metadata()["network_interface_ids"]; got != "eni-123,eni-456" {
		t.Fatalf("Compute Scan() network_interface_ids = %q, want eni-123,eni-456", got)
	}

	interfaces, err := NewAttachmentScanner(mockAttachmentClient{}, "us-east-1").scanNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatalf("scanNetworkInterfaces() unexpected error: %v", err)
	}
	if len(interfaces) != 1 {
		t.Fatalf("scanNetworkInterfaces() returned %d resources, want 1", len(interfaces))
	}
	if got := interfaces[0].Metadata()["instance_id"]; got != "i-123" {
		t.Fatalf("scanNetworkInterfaces() instance_id = %q, want i-123", got)
	}
}
