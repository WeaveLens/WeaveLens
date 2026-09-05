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
	aliases := make(map[string]string)
	for _, res := range resources {
		id := string(res.ID())
		resourceMap[id] = res
		aliases[id] = id
		if res.ARN() != "" {
			aliases[res.ARN()] = id
		}
		aliasKeys := []string{"arn", "dns_name", "topic_arn"}
		switch res.Type() {
		case "KMSKey":
			aliasKeys = append(aliasKeys, "key_arn")
		case "StepFunction":
			aliasKeys = append(aliasKeys, "state_machine_arn")
		case "TargetGroup":
			aliasKeys = append(aliasKeys, "target_group_arn")
		case "Listener":
			aliasKeys = append(aliasKeys, "listener_arn")
		}
		for _, key := range aliasKeys {
			if value := strings.TrimSuffix(strings.TrimSpace(res.Metadata()[key]), "."); value != "" {
				aliases[value] = id
			}
		}
	}

	for _, res := range resources {
		for _, rule := range metadataRules {
			addMetadataRelations(&relationships, res, resourceMap, aliases, rule.MetadataKey, rule.RelationType, false)
		}
		for _, rule := range typeRules {
			if string(res.Type()) == rule.SourceType {
				reverse := rule.RelationType == relationship.RelationshipContains
				addMetadataRelations(&relationships, res, resourceMap, aliases, rule.MetadataKey, rule.RelationType, reverse)
			}
		}
	}

	return relationships, nil
}

func addMetadataRelations(
	relationships *[]*relationship.Relationship,
	source *resource.Resource,
	resourceMap map[string]*resource.Resource,
	aliases map[string]string,
	metadataKey string,
	relationType relationship.RelationshipType,
	reverse bool,
) {
	ids := strings.Split(source.Metadata()[metadataKey], ",")
	seen := make(map[string]struct{}, len(ids))
	for _, targetID := range ids {
		targetID = strings.TrimSuffix(strings.TrimSpace(targetID), ".")
		if targetID == "" {
			continue
		}
		canonicalTargetID, exists := aliases[targetID]
		if !exists || canonicalTargetID == string(source.ID()) {
			continue
		}
		if _, exists := resourceMap[canonicalTargetID]; !exists {
			continue
		}
		if _, duplicate := seen[canonicalTargetID]; duplicate {
			continue
		}
		seen[canonicalTargetID] = struct{}{}
		sourceID := string(source.ID())
		if reverse {
			sourceID, canonicalTargetID = canonicalTargetID, sourceID
		}
		rel, err := relationship.NewRelationship(
			relationship.RelationshipID(fmt.Sprintf("rel-%s-%s-%s", sourceID, metadataKey, canonicalTargetID)),
			sourceID,
			canonicalTargetID,
			relationType,
		)
		if err == nil {
			*relationships = append(*relationships, rel)
		}
	}
}
