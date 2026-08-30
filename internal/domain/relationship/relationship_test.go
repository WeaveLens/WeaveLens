package relationship_test

import (
	"testing"

	"github.com/elip/WeaveLens/internal/domain/relationship"
)

func TestNewRelationship(t *testing.T) {
	tests := []struct {
		name      string
		id        relationship.RelationshipID
		sourceID  string
		targetID  string
		typ       relationship.RelationshipType
		opts      []relationship.RelationshipOption
		wantErr   bool
	}{
		{
			name:     "valid relationship",
			id:       "rel-1",
			sourceID: "vpc-1",
			targetID: "subnet-1",
			typ:      relationship.RelationshipContains,
			wantErr:  false,
		},
		{
			name:      "empty id",
			id:        "",
			sourceID:  "vpc-1",
			targetID:  "subnet-1",
			typ:       relationship.RelationshipContains,
			wantErr:   true,
		},
		{
			name:      "empty source id",
			id:        "rel-1",
			sourceID:  "",
			targetID:  "subnet-1",
			typ:       relationship.RelationshipContains,
			wantErr:   true,
		},
		{
			name:      "empty target id",
			id:        "rel-1",
			sourceID:  "vpc-1",
			targetID:  "",
			typ:       relationship.RelationshipContains,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := relationship.NewRelationship(tt.id, tt.sourceID, tt.targetID, tt.typ, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRelationship() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID() != tt.id {
					t.Errorf("NewRelationship() ID = %v, want %v", got.ID(), tt.id)
				}
				if got.SourceID() != tt.sourceID {
					t.Errorf("NewRelationship() SourceID = %v, want %v", got.SourceID(), tt.sourceID)
				}
				if got.TargetID() != tt.targetID {
					t.Errorf("NewRelationship() TargetID = %v, want %v", got.TargetID(), tt.targetID)
				}
				if got.Type() != tt.typ {
					t.Errorf("NewRelationship() Type = %v, want %v", got.Type(), tt.typ)
				}
			}
		})
	}
}

func TestRelationshipOptions(t *testing.T) {
	rel, err := relationship.NewRelationship(
		"rel-1",
		"vpc-1",
		"subnet-1",
		relationship.RelationshipContains,
		relationship.WithRelationshipMetadata(map[string]string{"key": "value"}),
	)
	if err != nil {
		t.Fatalf("NewRelationship() error = %v", err)
	}

	meta := rel.Metadata()
	if meta["key"] != "value" {
		t.Errorf("Metadata['key'] = %v, want value", meta["key"])
	}
}
