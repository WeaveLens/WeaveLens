package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("EC2", func(c *client.Clients, region string) Scanner { return NewComputeScanner(c.EC2, region) })
}

type ComputeScanner struct {
	client EC2ScannerAPI
	region string
}

func NewComputeScanner(client EC2ScannerAPI, region string) *ComputeScanner {
	return &ComputeScanner{client: client, region: region}
}

func (s *ComputeScanner) Name() string {
	return "EC2"
}

func (s *ComputeScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeInstancesPaginator(s.client, &ec2.DescribeInstancesInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "EC2", Err: ClassifyError(err)}
		}

		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				if instance.InstanceId == nil {
					continue
				}

				name := *instance.InstanceId
				tags := make(map[string]string)
				for _, tag := range instance.Tags {
					if tag.Key != nil && tag.Value != nil {
						tags[*tag.Key] = *tag.Value
						if *tag.Key == "Name" {
							name = *tag.Value
						}
					}
				}

				metadata := map[string]string{
					"instance_type": string(instance.InstanceType),
					"image_id":      safePtr(instance.ImageId),
				}
				if instance.State != nil {
					metadata["state"] = string(instance.State.Name)
				}
				if instance.VpcId != nil {
					metadata["vpc_id"] = *instance.VpcId
				}
				if instance.SubnetId != nil {
					metadata["subnet_id"] = *instance.SubnetId
				}
				var securityGroupIDs []string
				for _, group := range instance.SecurityGroups {
					if group.GroupId != nil {
						securityGroupIDs = append(securityGroupIDs, *group.GroupId)
					}
				}
				if len(securityGroupIDs) > 0 {
					metadata["security_group_ids"] = strings.Join(securityGroupIDs, ",")
				}
				var networkInterfaceIDs []string
				for _, networkInterface := range instance.NetworkInterfaces {
					if networkInterface.NetworkInterfaceId != nil {
						networkInterfaceIDs = append(networkInterfaceIDs, *networkInterface.NetworkInterfaceId)
					}
				}
				if len(networkInterfaceIDs) > 0 {
					metadata["network_interface_ids"] = strings.Join(networkInterfaceIDs, ",")
				}

				res, err := resource.NewResource(
					resource.ResourceID(*instance.InstanceId),
					resource.ResourceType("EC2"),
					resource.CategoryCompute,
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
	}
	return resources, nil
}

type EC2ScannerAPI interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}
