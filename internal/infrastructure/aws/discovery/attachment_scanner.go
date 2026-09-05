package discovery

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elip/WeaveLens/internal/domain/resource"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func init() {
	RegisterScanner("NetworkAttachments", func(c *client.Clients, region string) Scanner { return NewAttachmentScanner(c.EC2Attachments, region) })
}

// AttachmentScanner discovers network objects whose primary purpose is joining
// VPCs or routing traffic between them.
type AttachmentScanner struct {
	client client.EC2AttachmentAPI
	region string
}

func NewAttachmentScanner(client client.EC2AttachmentAPI, region string) *AttachmentScanner {
	return &AttachmentScanner{client: client, region: region}
}

func (s *AttachmentScanner) Name() string { return "NetworkAttachments" }

func (s *AttachmentScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	resources, err := s.scanVPCEndpoints(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "VPCEndpoint", Err: ClassifyError(err)}
	}
	transit, err := s.scanTransitGatewayAttachments(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "TransitGatewayAttachment", Err: ClassifyError(err)}
	}
	peering, err := s.scanVpcPeeringConnections(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "VPCPeering", Err: ClassifyError(err)}
	}
	interfaces, err := s.scanNetworkInterfaces(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "NetworkInterface", Err: ClassifyError(err)}
	}
	volumes, err := s.scanVolumes(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "EBS", Err: ClassifyError(err)}
	}
	return append(resources, append(transit, append(peering, append(interfaces, volumes...)...)...)...), nil
}

func (s *AttachmentScanner) scanVPCEndpoints(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeVpcEndpointsPaginator(s.client, &ec2.DescribeVpcEndpointsInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, endpoint := range page.VpcEndpoints {
			if endpoint.VpcEndpointId == nil {
				continue
			}
			metadata := map[string]string{
				"vpc_id":        safePtr(endpoint.VpcId),
				"service_name":  safePtr(endpoint.ServiceName),
				"endpoint_type": string(endpoint.VpcEndpointType),
			}
			if len(endpoint.SubnetIds) > 0 {
				metadata["subnet_ids"] = strings.Join(endpoint.SubnetIds, ",")
			}
			if len(endpoint.RouteTableIds) > 0 {
				metadata["route_table_ids"] = strings.Join(endpoint.RouteTableIds, ",")
			}
			var groupIDs []string
			for _, group := range endpoint.Groups {
				if group.GroupId != nil {
					groupIDs = append(groupIDs, *group.GroupId)
				}
			}
			if len(groupIDs) > 0 {
				metadata["security_group_ids"] = strings.Join(groupIDs, ",")
			}
			name := *endpoint.VpcEndpointId
			res, err := resource.NewResource(resource.ResourceID(name), resource.ResourceType("VPCEndpoint"), resource.CategoryNetwork, name, resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func (s *AttachmentScanner) scanTransitGatewayAttachments(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeTransitGatewayAttachmentsPaginator(s.client, &ec2.DescribeTransitGatewayAttachmentsInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, attachment := range page.TransitGatewayAttachments {
			if attachment.TransitGatewayAttachmentId == nil {
				continue
			}
			metadata := map[string]string{
				"transit_gateway_id": safePtr(attachment.TransitGatewayId),
				"resource_id":        safePtr(attachment.ResourceId),
				"resource_type":      string(attachment.ResourceType),
				"state":              string(attachment.State),
			}
			name := *attachment.TransitGatewayAttachmentId
			res, err := resource.NewResource(resource.ResourceID(name), resource.ResourceType("TransitGatewayAttachment"), resource.CategoryNetwork, name, resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func (s *AttachmentScanner) scanVpcPeeringConnections(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeVpcPeeringConnectionsPaginator(s.client, &ec2.DescribeVpcPeeringConnectionsInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, peering := range page.VpcPeeringConnections {
			if peering.VpcPeeringConnectionId == nil {
				continue
			}
			var vpcIDs []string
			if peering.RequesterVpcInfo != nil && peering.RequesterVpcInfo.VpcId != nil {
				vpcIDs = append(vpcIDs, *peering.RequesterVpcInfo.VpcId)
			}
			if peering.AccepterVpcInfo != nil && peering.AccepterVpcInfo.VpcId != nil {
				vpcIDs = append(vpcIDs, *peering.AccepterVpcInfo.VpcId)
			}
			metadata := map[string]string{"vpc_ids": strings.Join(vpcIDs, ","), "status": string(peering.Status.Code)}
			name := *peering.VpcPeeringConnectionId
			res, err := resource.NewResource(resource.ResourceID(name), resource.ResourceType("VPCPeering"), resource.CategoryNetwork, name, resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func (s *AttachmentScanner) scanNetworkInterfaces(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeNetworkInterfacesPaginator(s.client, &ec2.DescribeNetworkInterfacesInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, iface := range page.NetworkInterfaces {
			if iface.NetworkInterfaceId == nil {
				continue
			}
			metadata := map[string]string{"vpc_id": safePtr(iface.VpcId), "subnet_id": safePtr(iface.SubnetId)}
			if iface.Attachment != nil && iface.Attachment.InstanceId != nil {
				metadata["instance_id"] = *iface.Attachment.InstanceId
			}
			var groups []string
			for _, group := range iface.Groups {
				if group.GroupId != nil {
					groups = append(groups, *group.GroupId)
				}
			}
			if len(groups) > 0 {
				metadata["security_group_ids"] = strings.Join(groups, ",")
			}
			name := *iface.NetworkInterfaceId
			res, err := resource.NewResource(resource.ResourceID(name), resource.ResourceType("NetworkInterface"), resource.CategoryNetwork, name, resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func (s *AttachmentScanner) scanVolumes(ctx context.Context) ([]*resource.Resource, error) {
	paginator := ec2.NewDescribeVolumesPaginator(s.client, &ec2.DescribeVolumesInput{})
	var resources []*resource.Resource
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, volume := range page.Volumes {
			if volume.VolumeId == nil {
				continue
			}
			metadata := map[string]string{"state": string(volume.State), "availability_zone": safePtr(volume.AvailabilityZone)}
			var instanceIDs []string
			for _, attachment := range volume.Attachments {
				if attachment.InstanceId != nil {
					instanceIDs = append(instanceIDs, *attachment.InstanceId)
				}
			}
			if len(instanceIDs) > 0 {
				metadata["instance_ids"] = strings.Join(instanceIDs, ",")
			}
			name := *volume.VolumeId
			res, err := resource.NewResource(resource.ResourceID(name), resource.ResourceType("EBS"), resource.CategoryStorage, name, resource.WithMetadata(metadata), resource.WithRegion(s.region))
			if err == nil {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}
