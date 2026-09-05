package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("KMS", func(c *client.Clients, _ string) Scanner { return NewKMSScanner(c.KMS) })
}

type KMSScanner struct{ client KMSAPI }

func NewKMSScanner(client KMSAPI) *KMSScanner { return &KMSScanner{client: client} }
func (s *KMSScanner) Name() string            { return "KMS" }

func (s *KMSScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	aliases, err := s.listAliases(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "KMSAlias", Err: ClassifyError(err)}
	}
	paginator := kms.NewListKeysPaginator(s.client, &kms.ListKeysInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "KMS", Err: ClassifyError(err)}
		}
		for _, key := range page.Keys {
			if key.KeyId == nil {
				continue
			}
			metadata := map[string]string{"key_arn": safePtr(key.KeyArn)}
			keyAliases := aliases[*key.KeyId]
			if len(keyAliases) > 0 {
				metadata["key_alias"] = strings.Join(keyAliases, ",")
			}
			policy, policyErr := s.client.GetKeyPolicy(ctx, &kms.GetKeyPolicyInput{KeyId: key.KeyId, PolicyName: aws.String("default")})
			if policyErr == nil && policy != nil && policy.Policy != nil {
				metadata["key_policy"] = *policy.Policy
			}
			name := *key.KeyId
			if len(keyAliases) > 0 {
				name = keyAliases[0]
			}
			keyRes, resErr := resource.NewResource(resource.ResourceID(*key.KeyId), resource.ResourceType("KMSKey"), resource.CategorySecurity, name, resource.WithARN(safePtr(key.KeyArn)), resource.WithMetadata(metadata))
			if resErr == nil {
				resources = append(resources, keyRes)
			}
			for _, alias := range keyAliases {
				aliasRes, aliasErr := resource.NewResource(resource.ResourceID(alias), resource.ResourceType("KMSAlias"), resource.CategorySecurity, alias, resource.WithMetadata(map[string]string{"key_arn": safePtr(key.KeyArn), "key_alias": alias}))
				if aliasErr == nil {
					resources = append(resources, aliasRes)
				}
			}
		}
	}
	return resources, nil
}

func (s *KMSScanner) listAliases(ctx context.Context) (map[string][]string, error) {
	result := make(map[string][]string)
	paginator := kms.NewListAliasesPaginator(s.client, &kms.ListAliasesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, alias := range page.Aliases {
			if alias.TargetKeyId != nil && alias.AliasName != nil {
				result[*alias.TargetKeyId] = append(result[*alias.TargetKeyId], *alias.AliasName)
			}
		}
	}
	return result, nil
}

type KMSAPI interface {
	ListKeys(context.Context, *kms.ListKeysInput, ...func(*kms.Options)) (*kms.ListKeysOutput, error)
	ListAliases(context.Context, *kms.ListAliasesInput, ...func(*kms.Options)) (*kms.ListAliasesOutput, error)
	GetKeyPolicy(context.Context, *kms.GetKeyPolicyInput, ...func(*kms.Options)) (*kms.GetKeyPolicyOutput, error)
}
