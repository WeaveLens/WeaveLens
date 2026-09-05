package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("Elasticache", func(c *client.Clients, region string) Scanner { return NewElasticacheScanner(c.Elasticache, region) })
}

type ElasticacheScanner struct {
	client ElasticacheAPI
	region string
}

func NewElasticacheScanner(client ElasticacheAPI, region string) *ElasticacheScanner {
	return &ElasticacheScanner{client: client, region: region}
}

func (s *ElasticacheScanner) Name() string {
	return "Elasticache"
}

func (s *ElasticacheScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	var resources []*resource.Resource

	clusters, err := s.scanCacheClusters(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "ElasticacheCluster", Err: err}
	}
	resources = append(resources, clusters...)

	replicationGroups, err := s.scanReplicationGroups(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "ElasticacheReplicationGroup", Err: err}
	}
	resources = append(resources, replicationGroups...)

	return resources, nil
}

func (s *ElasticacheScanner) scanCacheClusters(ctx context.Context) ([]*resource.Resource, error) {
	paginator := elasticache.NewDescribeCacheClustersPaginator(s.client, &elasticache.DescribeCacheClustersInput{
		ShowCacheClustersNotInReplicationGroups: aws.Bool(true),
	})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, cluster := range page.CacheClusters {
			if cluster.CacheClusterId == nil {
				continue
			}

			name := safePtr(cluster.CacheClusterId)
			metadata := map[string]string{}
			if cluster.ARN != nil {
				metadata["arn"] = *cluster.ARN
			}
			if cluster.CacheNodeType != nil {
				metadata["node_type"] = *cluster.CacheNodeType
			}
			if cluster.Engine != nil {
				metadata["engine"] = *cluster.Engine
			}
			if cluster.EngineVersion != nil {
				metadata["engine_version"] = *cluster.EngineVersion
			}
			if cluster.CacheClusterStatus != nil {
				metadata["status"] = *cluster.CacheClusterStatus
			}
			if cluster.NumCacheNodes != nil {
				metadata["node_count"] = fmt.Sprintf("%d", *cluster.NumCacheNodes)
			}
			if cluster.PreferredAvailabilityZone != nil {
				metadata["availability_zone"] = *cluster.PreferredAvailabilityZone
			}

			tags := make(map[string]string)
			if arn := safePtr(cluster.ARN); arn != "" {
				tags = fetchElasticacheTags(ctx, s.client, arn)
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("ElasticacheCluster"),
				resource.CategoryDatabase,
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

func (s *ElasticacheScanner) scanReplicationGroups(ctx context.Context) ([]*resource.Resource, error) {
	paginator := elasticache.NewDescribeReplicationGroupsPaginator(s.client, &elasticache.DescribeReplicationGroupsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, rg := range page.ReplicationGroups {
			if rg.ReplicationGroupId == nil {
				continue
			}

			name := safePtr(rg.ReplicationGroupId)
			metadata := map[string]string{}
			if rg.ARN != nil {
				metadata["arn"] = *rg.ARN
			}
			if rg.Status != nil {
				metadata["status"] = *rg.Status
			}
			if rg.Engine != nil {
				metadata["engine"] = *rg.Engine
			}
			if rg.CacheNodeType != nil {
				metadata["node_type"] = *rg.CacheNodeType
			}
			if len(rg.NodeGroups) > 0 {
				metadata["node_group_count"] = fmt.Sprintf("%d", len(rg.NodeGroups))
			}
			if rg.AutomaticFailover != "" {
				metadata["automatic_failover"] = string(rg.AutomaticFailover)
			}
			if rg.MultiAZ != "" {
				metadata["multi_az"] = string(rg.MultiAZ)
			}

			tags := make(map[string]string)
			if arn := safePtr(rg.ARN); arn != "" {
				tags = fetchElasticacheTags(ctx, s.client, arn)
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("ElasticacheReplicationGroup"),
				resource.CategoryDatabase,
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

func fetchElasticacheTags(ctx context.Context, client ElasticacheAPI, arn string) map[string]string {
	tags := make(map[string]string)
	output, err := client.ListTagsForResource(ctx, &elasticache.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil || output == nil {
		return tags
	}
	for _, tag := range output.TagList {
		if tag.Key != nil && tag.Value != nil {
			tags[*tag.Key] = *tag.Value
		}
	}
	return tags
}

type ElasticacheAPI interface {
	DescribeCacheClusters(ctx context.Context, params *elasticache.DescribeCacheClustersInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error)
	DescribeReplicationGroups(ctx context.Context, params *elasticache.DescribeReplicationGroupsInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error)
	ListTagsForResource(ctx context.Context, params *elasticache.ListTagsForResourceInput, optFns ...func(*elasticache.Options)) (*elasticache.ListTagsForResourceOutput, error)
}
