package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/elip/WeaveLens/internal/domain/event"
)

type Publisher interface {
	PublishScanStarted(ctx context.Context, evt *event.ScanStartedEvent) error
	PublishScanCompleted(ctx context.Context, evt *event.ScanCompletedEvent) error
	PublishScanFailed(ctx context.Context, evt *event.ScanFailedEvent) error
	PublishResourceDiscovered(ctx context.Context, evt *event.ResourceDiscoveredEvent) error
	PublishRelationshipDiscovered(ctx context.Context, evt *event.RelationshipDiscoveredEvent) error
	PublishGraphCompleted(ctx context.Context, evt *event.GraphCompletedEvent) error
	Close() error
}

type Subscriber interface {
	SubscribeScanStarted(ctx context.Context, handler func(ctx context.Context, evt *event.ScanStartedEvent) error, opts ...SubscriberOption) (*Subscription, error)
	SubscribeScanCompleted(ctx context.Context, handler func(ctx context.Context, evt *event.ScanCompletedEvent) error, opts ...SubscriberOption) (*Subscription, error)
	SubscribeScanFailed(ctx context.Context, handler func(ctx context.Context, evt *event.ScanFailedEvent) error, opts ...SubscriberOption) (*Subscription, error)
	SubscribeResourceDiscovered(ctx context.Context, handler func(ctx context.Context, evt *event.ResourceDiscoveredEvent) error, opts ...SubscriberOption) (*Subscription, error)
	SubscribeRelationshipDiscovered(ctx context.Context, handler func(ctx context.Context, evt *event.RelationshipDiscoveredEvent) error, opts ...SubscriberOption) (*Subscription, error)
	SubscribeGraphCompleted(ctx context.Context, handler func(ctx context.Context, evt *event.GraphCompletedEvent) error, opts ...SubscriberOption) (*Subscription, error)
}

type EventBus struct {
	publisher  Publisher
	subscriber Subscriber
}

func NewEventBus(publisher Publisher, subscriber Subscriber) *EventBus {
	return &EventBus{
		publisher:  publisher,
		subscriber: subscriber,
	}
}

func (b *EventBus) PublishScanStarted(ctx context.Context, evt *event.ScanStartedEvent) error {
	return b.publisher.PublishScanStarted(ctx, evt)
}

func (b *EventBus) PublishScanCompleted(ctx context.Context, evt *event.ScanCompletedEvent) error {
	return b.publisher.PublishScanCompleted(ctx, evt)
}

func (b *EventBus) PublishScanFailed(ctx context.Context, evt *event.ScanFailedEvent) error {
	return b.publisher.PublishScanFailed(ctx, evt)
}

func (b *EventBus) PublishResourceDiscovered(ctx context.Context, evt *event.ResourceDiscoveredEvent) error {
	return b.publisher.PublishResourceDiscovered(ctx, evt)
}

func (b *EventBus) PublishRelationshipDiscovered(ctx context.Context, evt *event.RelationshipDiscoveredEvent) error {
	return b.publisher.PublishRelationshipDiscovered(ctx, evt)
}

func (b *EventBus) PublishGraphCompleted(ctx context.Context, evt *event.GraphCompletedEvent) error {
	return b.publisher.PublishGraphCompleted(ctx, evt)
}

func (b *EventBus) SubscribeScanStarted(ctx context.Context, handler func(ctx context.Context, evt *event.ScanStartedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return b.subscriber.SubscribeScanStarted(ctx, handler, opts...)
}

func (b *EventBus) SubscribeScanCompleted(ctx context.Context, handler func(ctx context.Context, evt *event.ScanCompletedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return b.subscriber.SubscribeScanCompleted(ctx, handler, opts...)
}

func (b *EventBus) SubscribeScanFailed(ctx context.Context, handler func(ctx context.Context, evt *event.ScanFailedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return b.subscriber.SubscribeScanFailed(ctx, handler, opts...)
}

func (b *EventBus) SubscribeResourceDiscovered(ctx context.Context, handler func(ctx context.Context, evt *event.ResourceDiscoveredEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return b.subscriber.SubscribeResourceDiscovered(ctx, handler, opts...)
}

func (b *EventBus) SubscribeRelationshipDiscovered(ctx context.Context, handler func(ctx context.Context, evt *event.RelationshipDiscoveredEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return b.subscriber.SubscribeRelationshipDiscovered(ctx, handler, opts...)
}

func (b *EventBus) SubscribeGraphCompleted(ctx context.Context, handler func(ctx context.Context, evt *event.GraphCompletedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return b.subscriber.SubscribeGraphCompleted(ctx, handler, opts...)
}

func (b *EventBus) Close() error {
	return b.publisher.Close()
}

type JetStreamPublisher struct {
	client *Client
}

func NewJetStreamPublisher(client *Client) *JetStreamPublisher {
	return &JetStreamPublisher{client: client}
}

func (p *JetStreamPublisher) PublishScanStarted(ctx context.Context, evt *event.ScanStartedEvent) error {
	data, err := encodeEvent(evt)
	if err != nil {
		return err
	}

	envelope := event.EventEnvelope{
		ID:       event.EventID(fmt.Sprintf("scan-started-%s", evt.ScanID)),
		Type:     event.EventTypeScanStarted,
		Version:  event.EventVersion,
		Occurred: time.Now(),
		ScanID:   evt.ScanID,
		Region:   evt.Region,
		Data:     data,
	}

	envelopeBytes, err := encodeEnvelope(envelope)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, "weavelens.scan.started.v1", envelopeBytes)
}

func (p *JetStreamPublisher) PublishScanCompleted(ctx context.Context, evt *event.ScanCompletedEvent) error {
	data, err := encodeEvent(evt)
	if err != nil {
		return err
	}

	envelope := event.EventEnvelope{
		ID:       event.EventID(fmt.Sprintf("scan-completed-%s", evt.ScanID)),
		Type:     event.EventTypeScanCompleted,
		Version:  event.EventVersion,
		Occurred: time.Now(),
		ScanID:   evt.ScanID,
		Data:     data,
	}

	envelopeBytes, err := encodeEnvelope(envelope)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, "weavelens.scan.completed.v1", envelopeBytes)
}

func (p *JetStreamPublisher) PublishScanFailed(ctx context.Context, evt *event.ScanFailedEvent) error {
	data, err := encodeEvent(evt)
	if err != nil {
		return err
	}

	envelope := event.EventEnvelope{
		ID:       event.EventID(fmt.Sprintf("scan-failed-%s", evt.ScanID)),
		Type:     event.EventTypeScanFailed,
		Version:  event.EventVersion,
		Occurred: time.Now(),
		ScanID:   evt.ScanID,
		Data:     data,
	}

	envelopeBytes, err := encodeEnvelope(envelope)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, "weavelens.scan.failed.v1", envelopeBytes)
}

func (p *JetStreamPublisher) PublishResourceDiscovered(ctx context.Context, evt *event.ResourceDiscoveredEvent) error {
	data, err := encodeEvent(evt)
	if err != nil {
		return err
	}

	envelope := event.EventEnvelope{
		ID:       event.EventID(fmt.Sprintf("resource-discovered-%s-%s", evt.ScanID, evt.Resource.ID)),
		Type:     event.EventTypeResourceDiscovered,
		Version:  event.EventVersion,
		Occurred: time.Now(),
		ScanID:   evt.ScanID,
		Region:   evt.Resource.Region,
		Data:     data,
	}

	envelopeBytes, err := encodeEnvelope(envelope)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, "weavelens.resource.discovered.v1", envelopeBytes)
}

func (p *JetStreamPublisher) PublishRelationshipDiscovered(ctx context.Context, evt *event.RelationshipDiscoveredEvent) error {
	data, err := encodeEvent(evt)
	if err != nil {
		return err
	}

	envelope := event.EventEnvelope{
		ID:       event.EventID(fmt.Sprintf("relationship-discovered-%s-%s", evt.ScanID, evt.Relationship.ID)),
		Type:     event.EventTypeRelationshipDiscovered,
		Version:  event.EventVersion,
		Occurred: time.Now(),
		ScanID:   evt.ScanID,
		Data:     data,
	}

	envelopeBytes, err := encodeEnvelope(envelope)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, "weavelens.relationship.discovered.v1", envelopeBytes)
}

func (p *JetStreamPublisher) PublishGraphCompleted(ctx context.Context, evt *event.GraphCompletedEvent) error {
	data, err := encodeEvent(evt)
	if err != nil {
		return err
	}

	envelope := event.EventEnvelope{
		ID:       event.EventID(fmt.Sprintf("graph-completed-%s", evt.ScanID)),
		Type:     event.EventTypeGraphCompleted,
		Version:  event.EventVersion,
		Occurred: time.Now(),
		ScanID:   evt.ScanID,
		Data:     data,
	}

	envelopeBytes, err := encodeEnvelope(envelope)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, "weavelens.graph.completed.v1", envelopeBytes)
}

func (p *JetStreamPublisher) Close() error {
	return p.client.Close()
}
