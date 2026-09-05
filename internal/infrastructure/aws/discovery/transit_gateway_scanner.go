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
	RegisterScanner("TransitGateway", func(c *client.Clients, region string) Scanner {
		return NewTransitGatewayScanner(c.TransitGateway, region)
	})
}

type TransitGatewayScanner struct {
	client TransitGatewayAPI
	region string
}

func NewTransitGatewayScanner(client TransitGatewayAPI, region string) *TransitGatewayScanner {
	return &TransitGatewayScanner{client: client, region: region}
}
func (s *TransitGatewayScanner) Name() string { return "TransitGateway" }

func (s *TransitGatewayScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	vpcIDs, err := s.vpcIDsByTransitGateway(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "TransitGatewayVpcAttachment", Err: ClassifyError(err)}
	}
	paginator := ec2.NewDescribeTransitGatewaysPaginator(s.client, &ec2.DescribeTransitGatewaysInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "TransitGateway", Err: ClassifyError(err)}
		}
		for _, gateway := range page.TransitGateways {
			if gateway.TransitGatewayId == nil {
				continue
			}
			name, tags := extractTags(gateway.Tags)
			if name == "" {
				name = *gateway.TransitGatewayId
			}
			metadata := map[string]string{"transit_gateway_id": *gateway.TransitGatewayId, "state": string(gateway.State)}
			if ids := vpcIDs[*gateway.TransitGatewayId]; len(ids) > 0 {
				metadata["vpc_id"] = strings.Join(ids, ",")
			}
			res, err := resource.NewResource(resource.ResourceID(*gateway.TransitGatewayId), resource.ResourceType("TransitGateway"), resource.CategoryNetwork, name, resource.WithMetadata(metadata), resource.WithTags(tags), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func (s *TransitGatewayScanner) vpcIDsByTransitGateway(ctx context.Context) (map[string][]string, error) {
	result := make(map[string][]string)
	paginator := ec2.NewDescribeTransitGatewayVpcAttachmentsPaginator(s.client, &ec2.DescribeTransitGatewayVpcAttachmentsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, attachment := range page.TransitGatewayVpcAttachments {
			if attachment.TransitGatewayId != nil && attachment.VpcId != nil {
				result[*attachment.TransitGatewayId] = append(result[*attachment.TransitGatewayId], *attachment.VpcId)
			}
		}
	}
	return result, nil
}

type TransitGatewayAPI interface {
	DescribeTransitGateways(context.Context, *ec2.DescribeTransitGatewaysInput, ...func(*ec2.Options)) (*ec2.DescribeTransitGatewaysOutput, error)
	DescribeTransitGatewayVpcAttachments(context.Context, *ec2.DescribeTransitGatewayVpcAttachmentsInput, ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayVpcAttachmentsOutput, error)
}
