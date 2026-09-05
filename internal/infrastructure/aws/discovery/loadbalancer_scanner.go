package discovery

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("ALB", func(c *client.Clients, region string) Scanner { return NewLoadBalancerScanner(c.ELBv2, region) })
}

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
	loadBalancers, loadBalancerARNs, err := s.scanLoadBalancers(ctx)
	if err != nil {
		return nil, err
	}
	targetGroups, err := s.scanTargetGroups(ctx)
	if err != nil {
		return nil, err
	}
	listeners, err := s.scanListeners(ctx, loadBalancerARNs)
	if err != nil {
		return nil, err
	}
	return append(loadBalancers, append(targetGroups, listeners...)...), nil
}

func (s *LoadBalancerScanner) scanLoadBalancers(ctx context.Context) ([]*resource.Resource, []string, error) {
	paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(s.client, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	var resources []*resource.Resource
	var loadBalancerARNs []string

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, &ScannerError{Scanner: "ALB", Err: ClassifyError(err)}
		}

		for _, lb := range page.LoadBalancers {
			if lb.LoadBalancerArn == nil || lb.LoadBalancerName == nil {
				continue
			}
			loadBalancerARNs = append(loadBalancerARNs, *lb.LoadBalancerArn)

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
			if len(lb.SecurityGroups) > 0 {
				metadata["security_group_ids"] = strings.Join(lb.SecurityGroups, ",")
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
	return resources, loadBalancerARNs, nil
}

func (s *LoadBalancerScanner) scanTargetGroups(ctx context.Context) ([]*resource.Resource, error) {
	paginator := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(s.client, &elasticloadbalancingv2.DescribeTargetGroupsInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "TargetGroup", Err: ClassifyError(err)}
		}
		for _, targetGroup := range page.TargetGroups {
			if targetGroup.TargetGroupArn == nil || targetGroup.TargetGroupName == nil {
				continue
			}
			metadata := map[string]string{
				"target_group_arn": *targetGroup.TargetGroupArn,
				"target_type":      string(targetGroup.TargetType),
			}
			if targetGroup.Port != nil {
				metadata["port"] = strconv.FormatInt(int64(*targetGroup.Port), 10)
			}
			if len(targetGroup.LoadBalancerArns) > 0 {
				metadata["load_balancer_arn"] = strings.Join(targetGroup.LoadBalancerArns, ",")
			}

			health, healthErr := s.client.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{TargetGroupArn: targetGroup.TargetGroupArn})
			if healthErr == nil && health != nil {
				instanceIDs := make(map[string]struct{})
				for _, description := range health.TargetHealthDescriptions {
					if description.Target != nil && description.Target.Id != nil {
						instanceIDs[*description.Target.Id] = struct{}{}
					}
				}
				if len(instanceIDs) > 0 {
					metadata["instance_ids"] = strings.Join(setValues(instanceIDs), ",")
				}
			}

			res, err := resource.NewResource(resource.ResourceID(*targetGroup.TargetGroupArn), resource.ResourceType("TargetGroup"), resource.CategoryNetwork, *targetGroup.TargetGroupName, resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func (s *LoadBalancerScanner) scanListeners(ctx context.Context, loadBalancerARNs []string) ([]*resource.Resource, error) {
	var resources []*resource.Resource
	for _, loadBalancerARN := range loadBalancerARNs {
		paginator := elasticloadbalancingv2.NewDescribeListenersPaginator(s.client, &elasticloadbalancingv2.DescribeListenersInput{LoadBalancerArn: &loadBalancerARN})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, &ScannerError{Scanner: "Listener", Err: ClassifyError(err)}
			}
			for _, listener := range page.Listeners {
				if listener.ListenerArn == nil {
					continue
				}
				metadata := map[string]string{
					"listener_arn":      *listener.ListenerArn,
					"load_balancer_arn": loadBalancerARN,
					"protocol":          string(listener.Protocol),
				}
				if listener.Port != nil {
					metadata["port"] = strconv.FormatInt(int64(*listener.Port), 10)
				}
				targetGroupARNs := make(map[string]struct{})
				for _, action := range listener.DefaultActions {
					if action.TargetGroupArn != nil {
						targetGroupARNs[*action.TargetGroupArn] = struct{}{}
					}
					if action.ForwardConfig != nil {
						for _, target := range action.ForwardConfig.TargetGroups {
							if target.TargetGroupArn != nil {
								targetGroupARNs[*target.TargetGroupArn] = struct{}{}
							}
						}
					}
				}
				if len(targetGroupARNs) > 0 {
					metadata["default_target_group_arn"] = strings.Join(setValues(targetGroupARNs), ",")
				}
				name := string(listener.Protocol)
				if listener.Port != nil {
					name = fmt.Sprintf("%s:%d", listener.Protocol, *listener.Port)
				}
				res, err := resource.NewResource(resource.ResourceID(*listener.ListenerArn), resource.ResourceType("Listener"), resource.CategoryNetwork, name, resource.WithMetadata(metadata), resource.WithRegion(s.region))
				if err == nil {
					resources = append(resources, res)
				}
			}
		}
	}
	return resources, nil
}

type ELBv2API interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error)
	DescribeListeners(ctx context.Context, params *elasticloadbalancingv2.DescribeListenersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeListenersOutput, error)
	DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error)
}
