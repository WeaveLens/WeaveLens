package discovery

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type mockELBv2Client struct{}

func (mockELBv2Client) DescribeLoadBalancers(context.Context, *elasticloadbalancingv2.DescribeLoadBalancersInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	return &elasticloadbalancingv2.DescribeLoadBalancersOutput{LoadBalancers: []elbtypes.LoadBalancer{{
		LoadBalancerArn:  aws.String("lb-arn"),
		LoadBalancerName: aws.String("load-balancer"),
		VpcId:            aws.String("vpc-123"),
	}}}, nil
}

func (mockELBv2Client) DescribeTargetGroups(context.Context, *elasticloadbalancingv2.DescribeTargetGroupsInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error) {
	return &elasticloadbalancingv2.DescribeTargetGroupsOutput{TargetGroups: []elbtypes.TargetGroup{{
		TargetGroupArn:   aws.String("tg-arn"),
		TargetGroupName:  aws.String("target-group"),
		LoadBalancerArns: []string{"lb-arn"},
		Port:             aws.Int32(8080),
		TargetType:       elbtypes.TargetTypeEnumInstance,
	}}}, nil
}

func (mockELBv2Client) DescribeListeners(context.Context, *elasticloadbalancingv2.DescribeListenersInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeListenersOutput, error) {
	return &elasticloadbalancingv2.DescribeListenersOutput{Listeners: []elbtypes.Listener{{
		ListenerArn: aws.String("listener-arn"),
		Port:        aws.Int32(443),
		Protocol:    elbtypes.ProtocolEnumHttps,
		DefaultActions: []elbtypes.Action{{
			TargetGroupArn: aws.String("tg-arn"),
			ForwardConfig:  &elbtypes.ForwardActionConfig{TargetGroups: []elbtypes.TargetGroupTuple{{TargetGroupArn: aws.String("tg-arn")}}},
		}},
	}}}, nil
}

func (mockELBv2Client) DescribeTargetHealth(context.Context, *elasticloadbalancingv2.DescribeTargetHealthInput, ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error) {
	return &elasticloadbalancingv2.DescribeTargetHealthOutput{TargetHealthDescriptions: []elbtypes.TargetHealthDescription{
		{Target: &elbtypes.TargetDescription{Id: aws.String("i-123")}},
		{Target: &elbtypes.TargetDescription{Id: aws.String("i-123")}},
	}}, nil
}

func TestLoadBalancerScanner_ScansRelatedResources(t *testing.T) {
	resources, err := NewLoadBalancerScanner(mockELBv2Client{}, "us-east-1").Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	if len(resources) != 3 {
		t.Fatalf("Scan() returned %d resources, want 3", len(resources))
	}

	byType := make(map[string]map[string]string, len(resources))
	for _, discovered := range resources {
		byType[string(discovered.Type())] = discovered.Metadata()
		if discovered.Region() != "us-east-1" {
			t.Errorf("resource %s region = %q, want us-east-1", discovered.ID(), discovered.Region())
		}
	}
	if got := byType["TargetGroup"]["instance_ids"]; got != "i-123" {
		t.Errorf("TargetGroup instance_ids = %q, want i-123", got)
	}
	if got := byType["TargetGroup"]["load_balancer_arn"]; got != "lb-arn" {
		t.Errorf("TargetGroup load_balancer_arn = %q, want lb-arn", got)
	}
	if got := byType["Listener"]["default_target_group_arn"]; got != "tg-arn" {
		t.Errorf("Listener default_target_group_arn = %q, want tg-arn", got)
	}
}
