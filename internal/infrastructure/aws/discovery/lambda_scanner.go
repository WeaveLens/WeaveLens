package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

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
			}
			if fn.Version != nil {
				metadata["version"] = *fn.Version
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
}
