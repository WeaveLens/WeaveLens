package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("EventBridge", func(c *client.Clients, region string) Scanner { return NewEventBridgeScanner(c.EventBridge, region) })
}

type EventBridgeScanner struct {
	client EventBridgeAPI
	region string
}

func NewEventBridgeScanner(client EventBridgeAPI, region string) *EventBridgeScanner {
	return &EventBridgeScanner{client: client, region: region}
}

func (s *EventBridgeScanner) Name() string {
	return "EventBridge"
}

func (s *EventBridgeScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	var resources []*resource.Resource
	input := &eventbridge.ListRulesInput{}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		output, err := s.client.ListRules(ctx, input)
		if err != nil {
			return nil, &ScannerError{Scanner: "EventBridge", Err: ClassifyError(err)}
		}

		for _, rule := range output.Rules {
			if rule.Name == nil {
				continue
			}

			name := safePtr(rule.Name)
			metadata := map[string]string{}
			if rule.Arn != nil {
				metadata["arn"] = *rule.Arn
			}
			if rule.Description != nil {
				metadata["description"] = *rule.Description
			}
			if rule.ScheduleExpression != nil {
				metadata["schedule"] = *rule.ScheduleExpression
			}
			if rule.EventPattern != nil {
				metadata["event_pattern"] = *rule.EventPattern
			}
			if rule.State != "" {
				metadata["state"] = string(rule.State)
			}
			if rule.EventBusName != nil {
				metadata["event_bus"] = *rule.EventBusName
			}
			if rule.ManagedBy != nil {
				metadata["managed_by"] = *rule.ManagedBy
			}
			targetOutput, targetErr := s.client.ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
				Rule:         rule.Name,
				EventBusName: rule.EventBusName,
			})
			if targetErr == nil && targetOutput != nil {
				var targetIDs []string
				for _, target := range targetOutput.Targets {
					if target.Arn != nil {
						targetIDs = append(targetIDs, resourceIDFromARN(*target.Arn))
					}
				}
				if len(targetIDs) > 0 {
					metadata["target_ids"] = strings.Join(targetIDs, ",")
				}
			}

			tags := make(map[string]string)
			if arn := safePtr(rule.Arn); arn != "" {
				tagOutput, tagErr := s.client.ListTagsForResource(ctx, &eventbridge.ListTagsForResourceInput{
					ResourceARN: aws.String(arn),
				})
				if tagErr == nil && tagOutput != nil {
					for _, tag := range tagOutput.Tags {
						if tag.Key != nil && tag.Value != nil {
							tags[*tag.Key] = *tag.Value
						}
					}
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("EventBridgeRule"),
				resource.CategoryIntegration,
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

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return resources, nil
}

type EventBridgeAPI interface {
	ListRules(ctx context.Context, params *eventbridge.ListRulesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListRulesOutput, error)
	DescribeRule(ctx context.Context, params *eventbridge.DescribeRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeRuleOutput, error)
	ListTagsForResource(ctx context.Context, params *eventbridge.ListTagsForResourceInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTagsForResourceOutput, error)
	ListTargetsByRule(ctx context.Context, params *eventbridge.ListTargetsByRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTargetsByRuleOutput, error)
}
