package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("Route53", func(c *client.Clients, _ string) Scanner { return NewRoute53Scanner(c.Route53) })
}

type Route53Scanner struct{ client Route53API }

func NewRoute53Scanner(client Route53API) *Route53Scanner { return &Route53Scanner{client: client} }
func (s *Route53Scanner) Name() string                    { return "Route53" }

func (s *Route53Scanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	paginator := route53.NewListHostedZonesPaginator(s.client, &route53.ListHostedZonesInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
		}
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "Route53", Err: ClassifyError(err)}
		}
		for _, zone := range page.HostedZones {
			if zone.Id == nil || zone.Name == nil {
				continue
			}
			zoneID := strings.TrimPrefix(*zone.Id, "/hostedzone/")
			zoneRes, resErr := resource.NewResource(resource.ResourceID(zoneID), resource.ResourceType("Route53HostedZone"), resource.CategoryNetwork, strings.TrimSuffix(*zone.Name, "."), resource.WithMetadata(map[string]string{"zone_id": zoneID}))
			if resErr == nil {
				resources = append(resources, zoneRes)
			}
			records, recordErr := s.scanRecords(ctx, zoneID)
			if recordErr != nil {
				return nil, recordErr
			}
			resources = append(resources, records...)
		}
	}
	return resources, nil
}

func (s *Route53Scanner) scanRecords(ctx context.Context, zoneID string) ([]*resource.Resource, error) {
	paginator := route53.NewListResourceRecordSetsPaginator(s.client, &route53.ListResourceRecordSetsInput{HostedZoneId: &zoneID})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ScannerError{Scanner: "Route53Record", Err: ClassifyError(err)}
		}
		for _, record := range page.ResourceRecordSets {
			recordType := string(record.Type)
			if record.Name == nil || (recordType != "A" && recordType != "AAAA" && recordType != "CNAME") {
				continue
			}
			name := strings.TrimSuffix(*record.Name, ".")
			metadata := map[string]string{"zone_id": zoneID, "record_name": name, "record_type": recordType}
			if record.AliasTarget != nil && record.AliasTarget.DNSName != nil {
				metadata["alias_target"] = strings.TrimSuffix(*record.AliasTarget.DNSName, ".")
			} else if recordType == "CNAME" && len(record.ResourceRecords) > 0 && record.ResourceRecords[0].Value != nil {
				metadata["alias_target"] = strings.TrimSuffix(*record.ResourceRecords[0].Value, ".")
			}
			id := fmt.Sprintf("%s:%s:%s", zoneID, name, recordType)
			res, err := resource.NewResource(resource.ResourceID(id), resource.ResourceType("Route53Record"), resource.CategoryNetwork, name, resource.WithMetadata(metadata))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

type Route53API interface {
	ListHostedZones(context.Context, *route53.ListHostedZonesInput, ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error)
	ListResourceRecordSets(context.Context, *route53.ListResourceRecordSetsInput, ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
}
