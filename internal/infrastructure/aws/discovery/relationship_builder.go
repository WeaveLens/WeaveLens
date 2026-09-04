package discovery

import (
	"fmt"
	"strings"

	"github.com/elip/WeaveLens/internal/domain/relationship"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type DefaultRelationshipBuilder struct{}

func NewRelationshipBuilder() *DefaultRelationshipBuilder {
	return &DefaultRelationshipBuilder{}
}

func (b *DefaultRelationshipBuilder) Build(resources []*resource.Resource) ([]*relationship.Relationship, error) {
	var relationships []*relationship.Relationship

	resourceMap := make(map[string]*resource.Resource)
	for _, res := range resources {
		resourceMap[string(res.ID())] = res
	}

	for _, res := range resources {
		addMetadataRelations(&relationships, res, resourceMap, "subnet_ids", relationship.RelationshipAssociatedWith)
		addMetadataRelations(&relationships, res, resourceMap, "security_group_ids", relationship.RelationshipAssociatedWith)
		addMetadataRelations(&relationships, res, resourceMap, "referenced_group_ids", relationship.RelationshipConnectsTo)
		addMetadataRelations(&relationships, res, resourceMap, "route_target_ids", relationship.RelationshipRoutesTo)
		addMetadataRelations(&relationships, res, resourceMap, "target_ids", relationship.RelationshipTargets)
		addMetadataRelations(&relationships, res, resourceMap, "subscription_target_ids", relationship.RelationshipTargets)
		addMetadataRelations(&relationships, res, resourceMap, "event_source_ids", relationship.RelationshipDependsOn)
		addMetadataRelations(&relationships, res, resourceMap, "iam_role_id", relationship.RelationshipDependsOn)
		addMetadataRelations(&relationships, res, resourceMap, "route_table_ids", relationship.RelationshipAssociatedWith)
		addMetadataRelations(&relationships, res, resourceMap, "vpc_ids", relationship.RelationshipConnectsTo)
		addMetadataRelations(&relationships, res, resourceMap, "resource_id", relationship.RelationshipAssociatedWith)

		switch res.Type() {
		case "Subnet":
			if vpcID, ok := res.Metadata()["vpc_id"]; ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", vpcID, res.ID())),
						vpcID,
						string(res.ID()),
						relationship.RelationshipContains,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "EC2":
			if subnetID, ok := res.Metadata()["subnet_id"]; ok && subnetID != "" {
				if _, exists := resourceMap[subnetID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", subnetID, res.ID())),
						subnetID,
						string(res.ID()),
						relationship.RelationshipContains,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "RouteTable":
			if vpcID, ok := res.Metadata()["vpc_id"]; ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", vpcID, res.ID())),
						vpcID,
						string(res.ID()),
						relationship.RelationshipContains,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "InternetGateway":
			if vpcID, ok := res.Metadata()["vpc_id"]; ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", vpcID, res.ID())),
						vpcID,
						string(res.ID()),
						relationship.RelationshipContains,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "NATGateway":
			if vpcID, ok := res.Metadata()["vpc_id"]; ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", vpcID, res.ID())),
						vpcID,
						string(res.ID()),
						relationship.RelationshipContains,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}
			if subnetID, ok := res.Metadata()["subnet_id"]; ok && subnetID != "" {
				if _, exists := resourceMap[subnetID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", subnetID, res.ID())),
						subnetID,
						string(res.ID()),
						relationship.RelationshipContains,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "SecurityGroup":
			if vpcID, ok := res.Metadata()["vpc_id"]; ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", vpcID, res.ID())),
						vpcID,
						string(res.ID()),
						relationship.RelationshipContains,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "ALB":
			if vpcID, ok := res.Metadata()["vpc_id"]; ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", vpcID, res.ID())),
						vpcID,
						string(res.ID()),
						relationship.RelationshipBelongsTo,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "RDS":
			if vpcID, ok := res.Metadata()["vpc_id"]; ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					rel, err := relationship.NewRelationship(
						relationship.RelationshipID(fmt.Sprintf("rel-%s-%s", vpcID, res.ID())),
						vpcID,
						string(res.ID()),
						relationship.RelationshipBelongsTo,
					)
					if err == nil {
						relationships = append(relationships, rel)
					}
				}
			}

		case "NetworkInterface":
			addSingleMetadataRelation(&relationships, res, resourceMap, "vpc_id", relationship.RelationshipAssociatedWith)
			addSingleMetadataRelation(&relationships, res, resourceMap, "subnet_id", relationship.RelationshipAssociatedWith)

		case "EBS":
			addMetadataRelations(&relationships, res, resourceMap, "instance_ids", relationship.RelationshipAssociatedWith)
		}
	}

	return relationships, nil
}

func addSingleMetadataRelation(
	relationships *[]*relationship.Relationship,
	source *resource.Resource,
	resourceMap map[string]*resource.Resource,
	metadataKey string,
	relationType relationship.RelationshipType,
) {
	addMetadataRelations(relationships, source, resourceMap, metadataKey, relationType)
}

func addMetadataRelations(
	relationships *[]*relationship.Relationship,
	source *resource.Resource,
	resourceMap map[string]*resource.Resource,
	metadataKey string,
	relationType relationship.RelationshipType,
) {
	ids := strings.Split(source.Metadata()[metadataKey], ",")
	seen := make(map[string]struct{}, len(ids))
	for _, targetID := range ids {
		targetID = strings.TrimSpace(targetID)
		if targetID == "" || targetID == string(source.ID()) {
			continue
		}
		if _, duplicate := seen[targetID]; duplicate {
			continue
		}
		seen[targetID] = struct{}{}
		if _, exists := resourceMap[targetID]; !exists {
			continue
		}
		rel, err := relationship.NewRelationship(
			relationship.RelationshipID(fmt.Sprintf("rel-%s-%s-%s", source.ID(), metadataKey, targetID)),
			string(source.ID()),
			targetID,
			relationType,
		)
		if err == nil {
			*relationships = append(*relationships, rel)
		}
	}
}
