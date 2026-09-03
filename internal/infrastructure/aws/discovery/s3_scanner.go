package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type S3Scanner struct {
	client S3API
	region string
}

func NewS3Scanner(client S3API) *S3Scanner {
	return &S3Scanner{client: client}
}

func (s *S3Scanner) Name() string {
	return "S3"
}

func (s *S3Scanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	output, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, &ScannerError{Scanner: "S3", Err: ClassifyError(err)}
	}

	var resources []*resource.Resource
	for _, bucket := range output.Buckets {
		if bucket.Name == nil {
			continue
		}

		metadata := map[string]string{}
		tags := make(map[string]string)

		if bucket.BucketArn != nil {
			metadata["arn"] = *bucket.BucketArn
		}
		if bucket.CreationDate != nil {
			metadata["creation_date"] = fmt.Sprintf("%d", bucket.CreationDate.Unix())
		}
		if bucket.BucketRegion != nil {
			metadata["bucket_region"] = *bucket.BucketRegion
		}

		tagOutput, tagErr := s.client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if tagErr == nil && tagOutput != nil {
			for _, tag := range tagOutput.TagSet {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}

		res, err := resource.NewResource(
			resource.ResourceID(*bucket.Name),
			resource.ResourceType("S3"),
			resource.CategoryStorage,
			*bucket.Name,
			resource.WithMetadata(metadata),
			resource.WithTags(tags),
		)
		if err != nil {
			continue
		}
		resources = append(resources, res)
	}

	return resources, nil
}

type S3API interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketTagging(ctx context.Context, params *s3.GetBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketTaggingOutput, error)
}
