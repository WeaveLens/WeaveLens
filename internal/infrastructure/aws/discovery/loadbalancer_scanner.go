package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type LoadBalancerScanner struct {
	client ELBv2API
	region string
}

func NewLoadBalancerScanner(client ELBv2API, region string) *LoadBalancerScanner {
	return &LoadBalancerScanner{client: client, region: region}
}

func (s *LoadBalancerScanner) Name() string {
	return "ALB"
}

func (s *LoadBalancerScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(s.client, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "ALB", Err: ClassifyError(err)}
		}

		for _, lb := range page.LoadBalancers {
			if lb.LoadBalancerArn == nil || lb.LoadBalancerName == nil {
				continue
			}

			metadata := map[string]string{
				"dns_name": safePtr(lb.DNSName),
				"type":     string(lb.Type),
				"scheme":   string(lb.Scheme),
			}
			if lb.VpcId != nil {
				metadata["vpc_id"] = *lb.VpcId
			}
			var subnetIDs []string
			for _, zone := range lb.AvailabilityZones {
				if zone.SubnetId != nil {
					subnetIDs = append(subnetIDs, *zone.SubnetId)
				}
			}
			if len(subnetIDs) > 0 {
				metadata["subnet_ids"] = strings.Join(subnetIDs, ",")
			}

			res, err := resource.NewResource(
				resource.ResourceID(*lb.LoadBalancerArn),
				resource.ResourceType("ALB"),
				resource.CategoryNetwork,
				*lb.LoadBalancerName,
				resource.WithMetadata(metadata),
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

type ELBv2API interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
}
