package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("Lambda", func(c *client.Clients, region string) Scanner { return NewLambdaScanner(c.Lambda, region) })
}

type LambdaScanner struct {
	client LambdaAPI
	region string
}

func NewLambdaScanner(client LambdaAPI, region string) *LambdaScanner {
	return &LambdaScanner{client: client, region: region}
}

func (s *LambdaScanner) Name() string {
	return "Lambda"
}

func (s *LambdaScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := lambda.NewListFunctionsPaginator(s.client, &lambda.ListFunctionsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "Lambda", Err: ClassifyError(err)}
		}

		for _, fn := range page.Functions {
			if fn.FunctionName == nil {
				continue
			}

			name := safePtr(fn.FunctionName)
			tags := make(map[string]string)

			metadata := map[string]string{
				"runtime": string(fn.Runtime),
				"state":   string(fn.State),
				"handler": safePtr(fn.Handler),
				"timeout": fmt.Sprintf("%d", fn.Timeout),
				"memory":  fmt.Sprintf("%d", fn.MemorySize),
			}
			if fn.FunctionArn != nil {
				metadata["arn"] = *fn.FunctionArn
			}
			metadata["code_size"] = fmt.Sprintf("%d", fn.CodeSize)
			if fn.LastModified != nil {
				metadata["last_modified"] = *fn.LastModified
			}
			if fn.Description != nil {
				metadata["description"] = *fn.Description
			}
			if fn.Role != nil {
				metadata["role"] = *fn.Role
				metadata["iam_role_id"] = resourceIDFromARN(*fn.Role)
			}
			if fn.Version != nil {
				metadata["version"] = *fn.Version
			}
			if fn.VpcConfig != nil {
				if fn.VpcConfig.VpcId != nil {
					metadata["vpc_id"] = *fn.VpcConfig.VpcId
				}
				if len(fn.VpcConfig.SubnetIds) > 0 {
					metadata["subnet_ids"] = strings.Join(fn.VpcConfig.SubnetIds, ",")
				}
				if len(fn.VpcConfig.SecurityGroupIds) > 0 {
					metadata["security_group_ids"] = strings.Join(fn.VpcConfig.SecurityGroupIds, ",")
				}
			}
			mappings, mappingErr := s.client.ListEventSourceMappings(ctx, &lambda.ListEventSourceMappingsInput{
				FunctionName: fn.FunctionName,
			})
			if mappingErr == nil && mappings != nil {
				var sourceIDs []string
				for _, mapping := range mappings.EventSourceMappings {
					if mapping.EventSourceArn != nil {
						sourceIDs = append(sourceIDs, resourceIDFromARN(*mapping.EventSourceArn))
					}
				}
				if len(sourceIDs) > 0 {
					metadata["event_source_ids"] = strings.Join(sourceIDs, ",")
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("Lambda"),
				resource.CategoryCompute,
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

type LambdaAPI interface {
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
	ListEventSourceMappings(ctx context.Context, params *lambda.ListEventSourceMappingsInput, optFns ...func(*lambda.Options)) (*lambda.ListEventSourceMappingsOutput, error)
}
