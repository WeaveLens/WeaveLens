package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("StepFunctions", func(c *client.Clients, region string) Scanner {
		return NewStepFunctionsScanner(c.StepFunctions, region)
	})
}

type StepFunctionsScanner struct {
	client StepFunctionsAPI
	region string
}

func NewStepFunctionsScanner(client StepFunctionsAPI, region string) *StepFunctionsScanner {
	return &StepFunctionsScanner{client: client, region: region}
}
func (s *StepFunctionsScanner) Name() string { return "StepFunctions" }

func (s *StepFunctionsScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := sfn.NewListStateMachinesPaginator(s.client, &sfn.ListStateMachinesInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "StepFunctions", Err: ClassifyError(err)}
		}
		for _, machine := range page.StateMachines {
			if machine.StateMachineArn == nil || machine.Name == nil {
				continue
			}
			metadata := map[string]string{"state_machine_arn": *machine.StateMachineArn, "type": string(machine.Type)}
			description, describeErr := s.client.DescribeStateMachine(ctx, &sfn.DescribeStateMachineInput{StateMachineArn: machine.StateMachineArn})
			if describeErr == nil && description != nil && description.Definition != nil {
				lambdaARNs, sfnARNs := definitionTargets(*description.Definition)
				if len(lambdaARNs) > 0 {
					metadata["target_lambda_arn"] = strings.Join(lambdaARNs, ",")
				}
				if len(sfnARNs) > 0 {
					metadata["target_sfn_arn"] = strings.Join(sfnARNs, ",")
				}
			}
			res, err := resource.NewResource(resource.ResourceID(*machine.StateMachineArn), resource.ResourceType("StepFunction"), resource.CategoryIntegration, *machine.Name, resource.WithARN(*machine.StateMachineArn), resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func definitionTargets(definition string) ([]string, []string) {
	var document any
	if json.Unmarshal([]byte(definition), &document) != nil {
		return nil, nil
	}
	lambdaSet := make(map[string]struct{})
	sfnSet := make(map[string]struct{})
	var walk func(any)
	walk = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		case string:
			if strings.HasPrefix(typed, "arn:") && strings.Contains(typed, ":lambda:") && strings.Contains(typed, ":function:") {
				lambdaSet[typed] = struct{}{}
			}
			if strings.HasPrefix(typed, "arn:") && strings.Contains(typed, ":states:") && strings.Contains(typed, ":stateMachine:") {
				sfnSet[typed] = struct{}{}
			}
		}
	}
	walk(document)
	return setValues(lambdaSet), setValues(sfnSet)
}

func setValues(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

type StepFunctionsAPI interface {
	ListStateMachines(context.Context, *sfn.ListStateMachinesInput, ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error)
	DescribeStateMachine(context.Context, *sfn.DescribeStateMachineInput, ...func(*sfn.Options)) (*sfn.DescribeStateMachineOutput, error)
}
