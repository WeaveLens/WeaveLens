package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("SQS", func(c *client.Clients, region string) Scanner { return NewSQSScanner(c.SQS, region) })
}

type SQSScanner struct {
	client SQSAPI
	region string
}

func NewSQSScanner(client SQSAPI, region string) *SQSScanner {
	return &SQSScanner{client: client, region: region}
}

func (s *SQSScanner) Name() string {
	return "SQS"
}

func (s *SQSScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := sqs.NewListQueuesPaginator(s.client, &sqs.ListQueuesInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "SQS", Err: ClassifyError(err)}
		}

		for _, queueUrl := range page.QueueUrls {
			if queueUrl == "" {
				continue
			}

			queueName := queueUrl
			lastIdx := -1
			for i := len(queueUrl) - 1; i >= 0; i-- {
				if queueUrl[i] == '/' {
					lastIdx = i
					break
				}
			}
			if lastIdx >= 0 && lastIdx < len(queueUrl)-1 {
				queueName = queueUrl[lastIdx+1:]
			}

			metadata := map[string]string{
				"queue_url": queueUrl,
			}

			attrs, attrErr := s.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
				QueueUrl:       &queueUrl,
				AttributeNames: []sqstypes.QueueAttributeName{"All"},
			})
			if attrErr == nil && attrs != nil {
				for k, v := range attrs.Attributes {
					metadata[string(k)] = v
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(queueName),
				resource.ResourceType("SQS"),
				resource.CategoryIntegration,
				queueName,
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

type SQSAPI interface {
	ListQueues(ctx context.Context, params *sqs.ListQueuesInput, optFns ...func(*sqs.Options)) (*sqs.ListQueuesOutput, error)
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}
