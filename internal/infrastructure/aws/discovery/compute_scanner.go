package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type ComputeScanner struct {
	client EC2ScannerAPI
}

func NewComputeScanner(client EC2ScannerAPI) *ComputeScanner {
	return &ComputeScanner{client: client}
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
				for _, tag := range instance.Tags {
					if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
						name = *tag.Value
						break
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

				res, err := resource.NewResource(
					resource.ResourceID(*instance.InstanceId),
					resource.ResourceType("EC2"),
					resource.CategoryCompute,
					name,
					resource.WithMetadata(metadata),
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
