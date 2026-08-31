package discovery

import (
	"fmt"

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
		}
	}

	return relationships, nil
}
