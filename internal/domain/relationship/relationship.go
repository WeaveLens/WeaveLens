package relationship

import "errors"

type RelationshipType string

const (
	RelationshipContains      RelationshipType = "contains"
	RelationshipBelongsTo     RelationshipType = "belongs_to"
	RelationshipConnectsTo    RelationshipType = "connects_to"
	RelationshipRoutesTo      RelationshipType = "routes_to"
	RelationshipTargets       RelationshipType = "targets"
	RelationshipDependsOn     RelationshipType = "depends_on"
	RelationshipAssociatedWith RelationshipType = "associated_with"
)

type RelationshipID string

type Relationship struct {
	id         RelationshipID
	sourceID   string
	targetID   string
	typ        RelationshipType
	metadata   map[string]string
}

type RelationshipOption func(*Relationship)

func WithRelationshipMetadata(metadata map[string]string) RelationshipOption {
	return func(r *Relationship) {
		r.metadata = metadata
	}
}

func NewRelationship(id RelationshipID, sourceID string, targetID string, typ RelationshipType, opts ...RelationshipOption) (*Relationship, error) {
	if id == "" {
		return nil, errors.New("relationship id must not be empty")
	}
	if sourceID == "" {
		return nil, errors.New("source id must not be empty")
	}
	if targetID == "" {
		return nil, errors.New("target id must not be empty")
	}

	r := &Relationship{
		id:       id,
		sourceID: sourceID,
		targetID: targetID,
		typ:      typ,
		metadata: make(map[string]string),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *Relationship) ID() RelationshipID {
	return r.id
}

func (r *Relationship) SourceID() string {
	return r.sourceID
}

func (r *Relationship) TargetID() string {
	return r.targetID
}

func (r *Relationship) Type() RelationshipType {
	return r.typ
}

func (r *Relationship) Metadata() map[string]string {
	return r.metadata
}
