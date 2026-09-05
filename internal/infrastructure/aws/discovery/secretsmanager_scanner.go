package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("SecretsManager", func(c *client.Clients, region string) Scanner { return NewSecretsManagerScanner(c.SecretsMgr, region) })
}

type SecretsManagerScanner struct {
	client SecretsManagerAPI
	region string
}

func NewSecretsManagerScanner(client SecretsManagerAPI, region string) *SecretsManagerScanner {
	return &SecretsManagerScanner{client: client, region: region}
}

func (s *SecretsManagerScanner) Name() string {
	return "SecretsManager"
}

func (s *SecretsManagerScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := secretsmanager.NewListSecretsPaginator(s.client, &secretsmanager.ListSecretsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "SecretsManager", Err: ClassifyError(err)}
		}

		for _, secret := range page.SecretList {
			if secret.Name == nil {
				continue
			}

			name := safePtr(secret.Name)
			metadata := map[string]string{}
			tags := make(map[string]string)

			if secret.ARN != nil {
				metadata["arn"] = *secret.ARN
			}
			if secret.Description != nil {
				metadata["description"] = *secret.Description
			}
			if secret.CreatedDate != nil {
				metadata["created_date"] = fmt.Sprintf("%d", secret.CreatedDate.Unix())
			}
			if secret.LastChangedDate != nil {
				metadata["last_changed"] = fmt.Sprintf("%d", secret.LastChangedDate.Unix())
			}
			if secret.LastAccessedDate != nil {
				metadata["last_accessed"] = fmt.Sprintf("%d", secret.LastAccessedDate.Unix())
			}
			if secret.KmsKeyId != nil {
				metadata["kms_key_id"] = *secret.KmsKeyId
			}
			if secret.OwningService != nil {
				metadata["owning_service"] = *secret.OwningService
			}

			for _, tag := range secret.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("Secret"),
				resource.CategorySecurity,
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

type SecretsManagerAPI interface {
	ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
	DescribeSecret(ctx context.Context, params *secretsmanager.DescribeSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DescribeSecretOutput, error)
}
