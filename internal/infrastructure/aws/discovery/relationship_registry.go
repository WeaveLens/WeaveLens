package discovery

import "github.com/elip/WeaveLens/internal/domain/relationship"

type MetadataRelationshipRule struct {
	MetadataKey  string
	RelationType relationship.RelationshipType
}

var metadataRules []MetadataRelationshipRule

func RegisterMetadataRule(key string, relType relationship.RelationshipType) {
	if key == "" || relType == "" {
		return
	}
	metadataRules = append(metadataRules, MetadataRelationshipRule{MetadataKey: key, RelationType: relType})
}

type TypeRelationshipRule struct {
	SourceType   string
	MetadataKey  string
	RelationType relationship.RelationshipType
}

var typeRules []TypeRelationshipRule

func RegisterTypeRule(sourceType, metadataKey string, relType relationship.RelationshipType) {
	if sourceType == "" || metadataKey == "" || relType == "" {
		return
	}
	typeRules = append(typeRules, TypeRelationshipRule{SourceType: sourceType, MetadataKey: metadataKey, RelationType: relType})
}

func init() {
	for _, rule := range []MetadataRelationshipRule{
		{"subnet_ids", relationship.RelationshipAssociatedWith},
		{"security_group_ids", relationship.RelationshipAssociatedWith},
		{"referenced_group_ids", relationship.RelationshipConnectsTo},
		{"route_target_ids", relationship.RelationshipRoutesTo},
		{"target_ids", relationship.RelationshipTargets},
		{"subscription_target_ids", relationship.RelationshipTargets},
		{"event_source_ids", relationship.RelationshipDependsOn},
		{"iam_role_id", relationship.RelationshipDependsOn},
		{"route_table_ids", relationship.RelationshipAssociatedWith},
		{"vpc_ids", relationship.RelationshipConnectsTo},
		{"resource_id", relationship.RelationshipAssociatedWith},
		{"network_interface_ids", relationship.RelationshipAssociatedWith},
		{"instance_id", relationship.RelationshipAssociatedWith},
		{"alias_target", relationship.RelationshipTargets},
		{"origin_domain", relationship.RelationshipTargets},
		{"target_lambda_arn", relationship.RelationshipTargets},
		{"target_sfn_arn", relationship.RelationshipTargets},
		{"kms_key_id", relationship.RelationshipAssociatedWith},
	} {
		RegisterMetadataRule(rule.MetadataKey, rule.RelationType)
	}

	for _, rule := range []TypeRelationshipRule{
		{"Subnet", "vpc_id", relationship.RelationshipContains},
		{"EC2", "subnet_id", relationship.RelationshipContains},
		{"RouteTable", "vpc_id", relationship.RelationshipContains},
		{"InternetGateway", "vpc_id", relationship.RelationshipContains},
		{"NATGateway", "vpc_id", relationship.RelationshipContains},
		{"NATGateway", "subnet_id", relationship.RelationshipContains},
		{"SecurityGroup", "vpc_id", relationship.RelationshipContains},
		{"ALB", "vpc_id", relationship.RelationshipBelongsTo},
		{"RDS", "vpc_id", relationship.RelationshipBelongsTo},
		{"NetworkInterface", "vpc_id", relationship.RelationshipAssociatedWith},
		{"NetworkInterface", "subnet_id", relationship.RelationshipAssociatedWith},
		{"EBS", "instance_ids", relationship.RelationshipAssociatedWith},
		{"Listener", "default_target_group_arn", relationship.RelationshipTargets},
		{"TargetGroup", "load_balancer_arn", relationship.RelationshipBelongsTo},
		{"TargetGroup", "instance_ids", relationship.RelationshipTargets},
		{"TransitGatewayAttachment", "transit_gateway_id", relationship.RelationshipAssociatedWith},
		{"TransitGateway", "vpc_id", relationship.RelationshipAssociatedWith},
		{"APIStage", "api_id", relationship.RelationshipBelongsTo},
		{"Route53Record", "zone_id", relationship.RelationshipBelongsTo},
		{"KMSAlias", "key_arn", relationship.RelationshipBelongsTo},
	} {
		RegisterTypeRule(rule.SourceType, rule.MetadataKey, rule.RelationType)
	}
}
