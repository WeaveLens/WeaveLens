package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type CloudWatchLogsScanner struct {
	client CloudWatchLogsAPI
	region string
}

func NewCloudWatchLogsScanner(client CloudWatchLogsAPI, region string) *CloudWatchLogsScanner {
	return &CloudWatchLogsScanner{client: client, region: region}
}

func (s *CloudWatchLogsScanner) Name() string {
	return "CloudWatchLogs"
}

func (s *CloudWatchLogsScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(s.client, &cloudwatchlogs.DescribeLogGroupsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "CloudWatchLogs", Err: ClassifyError(err)}
		}

		for _, group := range page.LogGroups {
			if group.LogGroupName == nil {
				continue
			}

			name := safePtr(group.LogGroupName)
			metadata := map[string]string{}

			if group.Arn != nil {
				metadata["arn"] = *group.Arn
			}
			if group.CreationTime != nil {
				metadata["created_at"] = fmt.Sprintf("%d", *group.CreationTime)
			}
			if group.RetentionInDays != nil {
				metadata["retention_days"] = fmt.Sprintf("%d", *group.RetentionInDays)
			}
			if group.StoredBytes != nil {
				metadata["stored_bytes"] = fmt.Sprintf("%d", *group.StoredBytes)
			}
			if group.MetricFilterCount != nil {
				metadata["metric_filter_count"] = fmt.Sprintf("%d", *group.MetricFilterCount)
			}
			if group.KmsKeyId != nil {
				metadata["kms_key_id"] = *group.KmsKeyId
			}

			tags := make(map[string]string)
			if arn := safePtr(group.Arn); arn != "" {
				tagOutput, tagErr := s.client.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
					ResourceArn: aws.String(arn),
				})
				if tagErr == nil && tagOutput != nil {
					for key, val := range tagOutput.Tags {
						tags[key] = val
					}
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("CloudWatchLogGroup"),
				resource.CategoryOther,
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
	return resources, nil
}

type CloudWatchLogsAPI interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	ListTagsForResource(ctx context.Context, params *cloudwatchlogs.ListTagsForResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListTagsForResourceOutput, error)
}
