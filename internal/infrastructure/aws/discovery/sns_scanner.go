package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("SNS", func(c *client.Clients, region string) Scanner { return NewSNSScanner(c.SNS, region) })
}

type SNSScanner struct {
	client SNSAPI
	region string
}

func NewSNSScanner(client SNSAPI, region string) *SNSScanner {
	return &SNSScanner{client: client, region: region}
}

func (s *SNSScanner) Name() string {
	return "SNS"
}

func (s *SNSScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := sns.NewListTopicsPaginator(s.client, &sns.ListTopicsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "SNS", Err: ClassifyError(err)}
		}

		for _, topic := range page.Topics {
			if topic.TopicArn == nil {
				continue
			}

			topicArn := *topic.TopicArn

			topicName := topicArn
			lastIdx := -1
			for i := len(topicArn) - 1; i >= 0; i-- {
				if topicArn[i] == ':' {
					lastIdx = i
					break
				}
			}
			if lastIdx >= 0 && lastIdx < len(topicArn)-1 {
				topicName = topicArn[lastIdx+1:]
			}

			metadata := map[string]string{
				"topic_arn": topicArn,
			}
			subscriptions, subErr := s.client.ListSubscriptionsByTopic(ctx, &sns.ListSubscriptionsByTopicInput{
				TopicArn: aws.String(topicArn),
			})
			if subErr == nil && subscriptions != nil {
				var targetIDs []string
				for _, subscription := range subscriptions.Subscriptions {
					if subscription.Endpoint != nil && strings.HasPrefix(*subscription.Endpoint, "arn:") {
						targetIDs = append(targetIDs, resourceIDFromARN(*subscription.Endpoint))
					}
				}
				if len(targetIDs) > 0 {
					metadata["subscription_target_ids"] = strings.Join(targetIDs, ",")
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(topicName),
				resource.ResourceType("SNS"),
				resource.CategoryIntegration,
				topicName,
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

type SNSAPI interface {
	ListTopics(ctx context.Context, params *sns.ListTopicsInput, optFns ...func(*sns.Options)) (*sns.ListTopicsOutput, error)
	ListSubscriptionsByTopic(ctx context.Context, params *sns.ListSubscriptionsByTopicInput, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsByTopicOutput, error)
}
