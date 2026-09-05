package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("APIGateway", func(c *client.Clients, region string) Scanner { return NewAPIGatewayScanner(c.APIGateway, region) })
}

type APIGatewayScanner struct {
	client APIGatewayAPI
	region string
}

func NewAPIGatewayScanner(client APIGatewayAPI, region string) *APIGatewayScanner {
	return &APIGatewayScanner{client: client, region: region}
}
func (s *APIGatewayScanner) Name() string { return "APIGateway" }

func (s *APIGatewayScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := apigateway.NewGetRestApisPaginator(s.client, &apigateway.GetRestApisInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "APIGateway", Err: ClassifyError(err)}
		}
		for _, api := range page.Items {
			if api.Id == nil || api.Name == nil {
				continue
			}
			lambdaARNs := s.integrationLambdaARNs(ctx, *api.Id)
			metadata := map[string]string{"api_id": *api.Id}
			if len(lambdaARNs) > 0 {
				metadata["target_lambda_arn"] = strings.Join(lambdaARNs, ",")
			}
			apiRes, resErr := resource.NewResource(resource.ResourceID(*api.Id), resource.ResourceType("APIGateway"), resource.CategoryIntegration, *api.Name, resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if resErr == nil {
				resources = append(resources, apiRes)
			}

			stages, stageErr := s.client.GetStages(ctx, &apigateway.GetStagesInput{RestApiId: api.Id})
			if stageErr != nil {
				return nil, &ScannerError{Scanner: "APIGatewayStage", Err: ClassifyError(stageErr)}
			}
			for _, stage := range stages.Item {
				if stage.StageName == nil {
					continue
				}
				stageMetadata := map[string]string{"api_id": *api.Id, "stage_name": *stage.StageName}
				if len(lambdaARNs) > 0 {
					stageMetadata["target_lambda_arn"] = strings.Join(lambdaARNs, ",")
				}
				id := fmt.Sprintf("%s:%s", *api.Id, *stage.StageName)
				stageRes, err := resource.NewResource(resource.ResourceID(id), resource.ResourceType("APIStage"), resource.CategoryIntegration, *stage.StageName, resource.WithMetadata(stageMetadata), resource.WithRegion(s.region))
				if err == nil {
					resources = append(resources, stageRes)
				}
			}
		}
	}
	return resources, nil
}

func (s *APIGatewayScanner) integrationLambdaARNs(ctx context.Context, apiID string) []string {
	paginator := apigateway.NewGetResourcesPaginator(s.client, &apigateway.GetResourcesInput{RestApiId: &apiID})
	seen := make(map[string]struct{})
	var result []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return result
		}
		for _, apiResource := range page.Items {
			if apiResource.Id == nil {
				continue
			}
			for method := range apiResource.ResourceMethods {
				integration, err := s.client.GetIntegration(ctx, &apigateway.GetIntegrationInput{RestApiId: &apiID, ResourceId: apiResource.Id, HttpMethod: &method})
				if err != nil || integration.Uri == nil {
					continue
				}
				if arn := lambdaARNFromIntegrationURI(*integration.Uri); arn != "" {
					if _, exists := seen[arn]; !exists {
						seen[arn] = struct{}{}
						result = append(result, arn)
					}
				}
			}
		}
	}
	return result
}

func lambdaARNFromIntegrationURI(uri string) string {
	const marker = "/functions/"
	start := strings.Index(uri, marker)
	if start < 0 {
		return ""
	}
	arn := uri[start+len(marker):]
	if end := strings.Index(arn, "/invocations"); end >= 0 {
		arn = arn[:end]
	}
	if strings.HasPrefix(arn, "arn:") {
		return arn
	}
	return ""
}

type APIGatewayAPI interface {
	GetRestApis(context.Context, *apigateway.GetRestApisInput, ...func(*apigateway.Options)) (*apigateway.GetRestApisOutput, error)
	GetStages(context.Context, *apigateway.GetStagesInput, ...func(*apigateway.Options)) (*apigateway.GetStagesOutput, error)
	GetResources(context.Context, *apigateway.GetResourcesInput, ...func(*apigateway.Options)) (*apigateway.GetResourcesOutput, error)
	GetIntegration(context.Context, *apigateway.GetIntegrationInput, ...func(*apigateway.Options)) (*apigateway.GetIntegrationOutput, error)
}
