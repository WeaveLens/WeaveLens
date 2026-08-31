package event

import "time"

type EventID string

type Event struct {
	ID        EventID
	Type      string
	Version   string
	Occurred  time.Time
	ScanID    string
	AccountID string
	Region    string
	Payload   interface{}
}

type EventEnvelope struct {
	ID        EventID
	Type      string
	Version   string
	Occurred  time.Time
	ScanID    string
	AccountID string
	Region    string
	Data      []byte
}

type ScanStartedEvent struct {
	ScanID string
	Region string
}

type ScanCompletedEvent struct {
	ScanID       string
	ResourceCount int
}

type ScanFailedEvent struct {
	ScanID  string
	Error   string
}

type ResourceDiscoveredEvent struct {
	ScanID    string
	Resource  Resource
}

type RelationshipDiscoveredEvent struct {
	ScanID       string
	Relationship Relationship
}

type GraphCompletedEvent struct {
	ScanID       string
	NodeCount    int
	EdgeCount    int
}

type Resource struct {
	ID         string
	ARN        string
	AccountID  string
	Region     string
	Type       string
	Category   string
	Name       string
	Metadata   map[string]string
}

type Relationship struct {
	ID       string
	SourceID string
	TargetID string
	Type     string
	Metadata map[string]string
}

const (
	EventTypeScanStarted         = "scan.started"
	EventTypeScanCompleted       = "scan.completed"
	EventTypeScanFailed          = "scan.failed"
	EventTypeResourceDiscovered  = "resource.discovered"
	EventTypeRelationshipDiscovered = "relationship.discovered"
	EventTypeGraphCompleted      = "graph.completed"
)

const EventVersion = "v1"
