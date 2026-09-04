package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type DatabaseScanner struct {
	client RDSAPI
	region string
}

func NewDatabaseScanner(client RDSAPI, region string) *DatabaseScanner {
	return &DatabaseScanner{client: client, region: region}
}

func (s *DatabaseScanner) Name() string {
	return "RDS"
}

func (s *DatabaseScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := rds.NewDescribeDBInstancesPaginator(s.client, &rds.DescribeDBInstancesInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "RDS", Err: ClassifyError(err)}
		}

		for _, db := range page.DBInstances {
			if db.DBInstanceIdentifier == nil {
				continue
			}

			tags := make(map[string]string)
			for _, tag := range db.TagList {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}

			metadata := map[string]string{
				"engine":         safePtr(db.Engine),
				"instance_class": safePtr(db.DBInstanceClass),
				"status":         safePtr(db.DBInstanceStatus),
			}
			if db.DBSubnetGroup != nil && db.DBSubnetGroup.VpcId != nil {
				metadata["vpc_id"] = *db.DBSubnetGroup.VpcId
			}
			if db.DBSubnetGroup != nil {
				var subnetIDs []string
				for _, subnet := range db.DBSubnetGroup.Subnets {
					if subnet.SubnetIdentifier != nil {
						subnetIDs = append(subnetIDs, *subnet.SubnetIdentifier)
					}
				}
				if len(subnetIDs) > 0 {
					metadata["subnet_ids"] = strings.Join(subnetIDs, ",")
				}
			}

			res, err := resource.NewResource(
				resource.ResourceID(*db.DBInstanceIdentifier),
				resource.ResourceType("RDS"),
				resource.CategoryDatabase,
				*db.DBInstanceIdentifier,
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

type RDSAPI interface {
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}
