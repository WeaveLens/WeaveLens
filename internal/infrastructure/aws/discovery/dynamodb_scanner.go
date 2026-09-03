package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type DynamoDBScanner struct {
	client DynamoDBAPI
	region string
}

func NewDynamoDBScanner(client DynamoDBAPI, region string) *DynamoDBScanner {
	return &DynamoDBScanner{client: client, region: region}
}

func (s *DynamoDBScanner) Name() string {
	return "DynamoDB"
}

func (s *DynamoDBScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := dynamodb.NewListTablesPaginator(s.client, &dynamodb.ListTablesInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "DynamoDB", Err: ClassifyError(err)}
		}

		for _, tableName := range page.TableNames {
			descOutput, descErr := s.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if descErr != nil {
				res, resErr := resource.NewResource(
					resource.ResourceID(tableName),
					resource.ResourceType("DynamoDBTable"),
					resource.CategoryDatabase,
					tableName,
					resource.WithMetadata(map[string]string{"error": "describe_table_failed"}),
					resource.WithRegion(s.region),
				)
				if resErr == nil {
					resources = append(resources, res)
				}
				continue
			}

			table := descOutput.Table
			if table == nil {
				continue
			}

			metadata := map[string]string{}
			if table.TableArn != nil {
				metadata["arn"] = *table.TableArn
			}
			if table.TableStatus != "" {
				metadata["status"] = string(table.TableStatus)
			}
			if table.CreationDateTime != nil {
				metadata["created_at"] = fmt.Sprintf("%d", table.CreationDateTime.Unix())
			}
			if table.BillingModeSummary != nil && table.BillingModeSummary.BillingMode != "" {
				metadata["billing_mode"] = string(table.BillingModeSummary.BillingMode)
			}
			if table.ProvisionedThroughput != nil {
				metadata["read_capacity"] = fmt.Sprintf("%d", table.ProvisionedThroughput.ReadCapacityUnits)
				metadata["write_capacity"] = fmt.Sprintf("%d", table.ProvisionedThroughput.WriteCapacityUnits)
			}
			if table.SSEDescription != nil && table.SSEDescription.Status != "" {
				metadata["sse_status"] = string(table.SSEDescription.Status)
			}

			res, err := resource.NewResource(
				resource.ResourceID(tableName),
				resource.ResourceType("DynamoDBTable"),
				resource.CategoryDatabase,
				tableName,
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

type DynamoDBAPI interface {
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
}
