package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("CloudFront", func(c *client.Clients, _ string) Scanner { return NewCloudFrontScanner(c.CloudFront) })
}

type CloudFrontScanner struct{ client CloudFrontAPI }

func NewCloudFrontScanner(client CloudFrontAPI) *CloudFrontScanner {
	return &CloudFrontScanner{client: client}
}
func (s *CloudFrontScanner) Name() string { return "CloudFront" }

func (s *CloudFrontScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := cloudfront.NewListDistributionsPaginator(s.client, &cloudfront.ListDistributionsInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "CloudFront", Err: ClassifyError(err)}
		}
		if page.DistributionList == nil {
			continue
		}
		for _, distribution := range page.DistributionList.Items {
			if distribution.Id == nil {
				continue
			}
			name := safePtr(distribution.Comment)
			if name == "" {
				name = *distribution.Id
			}
			metadata := map[string]string{"dns_name": safePtr(distribution.DomainName), "status": safePtr(distribution.Status)}
			var domains, originTypes []string
			if distribution.Origins != nil {
				for _, origin := range distribution.Origins.Items {
					domain := safePtr(origin.DomainName)
					if domain == "" {
						continue
					}
					domains = append(domains, domain)
					switch {
					case origin.S3OriginConfig != nil:
						originTypes = append(originTypes, "S3")
					case strings.Contains(domain, ".elb.amazonaws.com"):
						originTypes = append(originTypes, "ALB")
					default:
						originTypes = append(originTypes, "Custom")
					}
				}
			}
			metadata["origin_domain"] = strings.Join(domains, ",")
			metadata["origin_type"] = strings.Join(originTypes, ",")
			res, err := resource.NewResource(resource.ResourceID(*distribution.Id), resource.ResourceType("CloudFrontDistribution"), resource.CategoryNetwork, name, resource.WithMetadata(metadata))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

type CloudFrontAPI interface {
	ListDistributions(context.Context, *cloudfront.ListDistributionsInput, ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error)
}
