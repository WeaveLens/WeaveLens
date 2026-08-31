package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type LoadBalancerScanner struct {
	client ELBv2API
}

func NewLoadBalancerScanner(client ELBv2API) *LoadBalancerScanner {
	return &LoadBalancerScanner{client: client}
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

			res, err := resource.NewResource(
				resource.ResourceID(*lb.LoadBalancerArn),
				resource.ResourceType("ALB"),
				resource.CategoryNetwork,
				*lb.LoadBalancerName,
				resource.WithMetadata(metadata),
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
