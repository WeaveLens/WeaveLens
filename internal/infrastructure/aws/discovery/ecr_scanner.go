package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type ECRScanner struct {
	client ECRAPI
	region string
}

func NewECRScanner(client ECRAPI, region string) *ECRScanner {
	return &ECRScanner{client: client, region: region}
}

func (s *ECRScanner) Name() string {
	return "ECR"
}

func (s *ECRScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ecr.NewDescribeRepositoriesPaginator(s.client, &ecr.DescribeRepositoriesInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "ECR", Err: ClassifyError(err)}
		}

		for _, repo := range page.Repositories {
			if repo.RepositoryName == nil {
				continue
			}

			name := safePtr(repo.RepositoryName)
			metadata := map[string]string{}
			tags := make(map[string]string)

			if repo.RepositoryArn != nil {
				metadata["arn"] = *repo.RepositoryArn
			}
			if repo.RepositoryUri != nil {
				metadata["uri"] = *repo.RepositoryUri
			}
			if repo.CreatedAt != nil {
				metadata["created_at"] = fmt.Sprintf("%d", repo.CreatedAt.Unix())
			}
			if repo.ImageScanningConfiguration != nil {
				metadata["image_scanning"] = fmt.Sprintf("%t", repo.ImageScanningConfiguration.ScanOnPush)
			}
			if repo.ImageTagMutability != "" {
				metadata["image_tag_mutability"] = string(repo.ImageTagMutability)
			}

			arn := safePtr(repo.RepositoryArn)
			if arn != "" {
				tagOutput, tagErr := s.client.ListTagsForResource(ctx, &ecr.ListTagsForResourceInput{
					ResourceArn: repo.RepositoryArn,
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
				resource.ResourceType("ECRRepository"),
				resource.CategoryStorage,
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

type ECRAPI interface {
	DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
	ListTagsForResource(ctx context.Context, params *ecr.ListTagsForResourceInput, optFns ...func(*ecr.Options)) (*ecr.ListTagsForResourceOutput, error)
}
