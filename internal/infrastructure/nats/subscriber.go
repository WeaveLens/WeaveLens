package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/elip/WeaveLens/internal/domain/event"
	"github.com/nats-io/nats.go"
)

type JetStreamSubscriber struct {
	client *Client
	logger *slog.Logger
}

func NewJetStreamSubscriber(client *Client) *JetStreamSubscriber {
	logger := slog.Default()
	return &JetStreamSubscriber{
		client: client,
		logger: logger,
	}
}

func (s *JetStreamSubscriber) SubscribeScanStarted(ctx context.Context, handler func(ctx context.Context, evt *event.ScanStartedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return s.client.Subscribe("weavelens.scan.started.v1", func(msg *nats.Msg) error {
		start := time.Now()
		envelope := &event.EventEnvelope{
			ID:       event.EventID(msg.Header.Get("Nats-Message-Id")),
			Type:     msg.Subject,
			Data:     msg.Data,
			Occurred: time.Now(),
		}

		evt, err := decodeEvent(envelope, &event.ScanStartedEvent{})
		if err != nil {
			s.logger.Error("failed to decode event", "eventID", envelope.ID, "subject", envelope.Type, "error", err)
			return err
		}

		err = handler(ctx, evt.(*event.ScanStartedEvent))
		duration := time.Since(start)

		if err != nil {
			s.logger.Error("handler failed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ScanStartedEvent).ScanID, "duration", duration, "error", err)
		} else {
			s.logger.Info("event processed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ScanStartedEvent).ScanID, "duration", duration)
		}

		return err
	}, opts...)
}

func (s *JetStreamSubscriber) SubscribeScanCompleted(ctx context.Context, handler func(ctx context.Context, evt *event.ScanCompletedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return s.client.Subscribe("weavelens.scan.completed.v1", func(msg *nats.Msg) error {
		start := time.Now()
		envelope := &event.EventEnvelope{
			ID:       event.EventID(msg.Header.Get("Nats-Message-Id")),
			Type:     msg.Subject,
			Data:     msg.Data,
			Occurred: time.Now(),
		}

		evt, err := decodeEvent(envelope, &event.ScanCompletedEvent{})
		if err != nil {
			s.logger.Error("failed to decode event", "eventID", envelope.ID, "subject", envelope.Type, "error", err)
			return err
		}

		err = handler(ctx, evt.(*event.ScanCompletedEvent))
		duration := time.Since(start)

		if err != nil {
			s.logger.Error("handler failed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ScanCompletedEvent).ScanID, "duration", duration, "error", err)
		} else {
			s.logger.Info("event processed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ScanCompletedEvent).ScanID, "duration", duration)
		}

		return err
	}, opts...)
}

func (s *JetStreamSubscriber) SubscribeScanFailed(ctx context.Context, handler func(ctx context.Context, evt *event.ScanFailedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return s.client.Subscribe("weavelens.scan.failed.v1", func(msg *nats.Msg) error {
		start := time.Now()
		envelope := &event.EventEnvelope{
			ID:       event.EventID(msg.Header.Get("Nats-Message-Id")),
			Type:     msg.Subject,
			Data:     msg.Data,
			Occurred: time.Now(),
		}

		evt, err := decodeEvent(envelope, &event.ScanFailedEvent{})
		if err != nil {
			s.logger.Error("failed to decode event", "eventID", envelope.ID, "subject", envelope.Type, "error", err)
			return err
		}

		err = handler(ctx, evt.(*event.ScanFailedEvent))
		duration := time.Since(start)

		if err != nil {
			s.logger.Error("handler failed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ScanFailedEvent).ScanID, "duration", duration, "error", err)
		} else {
			s.logger.Info("event processed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ScanFailedEvent).ScanID, "duration", duration)
		}

		return err
	}, opts...)
}

func (s *JetStreamSubscriber) SubscribeResourceDiscovered(ctx context.Context, handler func(ctx context.Context, evt *event.ResourceDiscoveredEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return s.client.Subscribe("weavelens.resource.discovered.v1", func(msg *nats.Msg) error {
		start := time.Now()
		envelope := &event.EventEnvelope{
			ID:       event.EventID(msg.Header.Get("Nats-Message-Id")),
			Type:     msg.Subject,
			Data:     msg.Data,
			Occurred: time.Now(),
		}

		evt, err := decodeEvent(envelope, &event.ResourceDiscoveredEvent{})
		if err != nil {
			s.logger.Error("failed to decode event", "eventID", envelope.ID, "subject", envelope.Type, "error", err)
			return err
		}

		err = handler(ctx, evt.(*event.ResourceDiscoveredEvent))
		duration := time.Since(start)

		if err != nil {
			s.logger.Error("handler failed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ResourceDiscoveredEvent).ScanID, "duration", duration, "error", err)
		} else {
			s.logger.Info("event processed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.ResourceDiscoveredEvent).ScanID, "duration", duration)
		}

		return err
	}, opts...)
}

func (s *JetStreamSubscriber) SubscribeRelationshipDiscovered(ctx context.Context, handler func(ctx context.Context, evt *event.RelationshipDiscoveredEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return s.client.Subscribe("weavelens.relationship.discovered.v1", func(msg *nats.Msg) error {
		start := time.Now()
		envelope := &event.EventEnvelope{
			ID:       event.EventID(msg.Header.Get("Nats-Message-Id")),
			Type:     msg.Subject,
			Data:     msg.Data,
			Occurred: time.Now(),
		}

		evt, err := decodeEvent(envelope, &event.RelationshipDiscoveredEvent{})
		if err != nil {
			s.logger.Error("failed to decode event", "eventID", envelope.ID, "subject", envelope.Type, "error", err)
			return err
		}

		err = handler(ctx, evt.(*event.RelationshipDiscoveredEvent))
		duration := time.Since(start)

		if err != nil {
			s.logger.Error("handler failed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.RelationshipDiscoveredEvent).ScanID, "duration", duration, "error", err)
		} else {
			s.logger.Info("event processed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.RelationshipDiscoveredEvent).ScanID, "duration", duration)
		}

		return err
	}, opts...)
}

func (s *JetStreamSubscriber) SubscribeGraphCompleted(ctx context.Context, handler func(ctx context.Context, evt *event.GraphCompletedEvent) error, opts ...SubscriberOption) (*Subscription, error) {
	return s.client.Subscribe("weavelens.graph.completed.v1", func(msg *nats.Msg) error {
		start := time.Now()
		envelope := &event.EventEnvelope{
			ID:       event.EventID(msg.Header.Get("Nats-Message-Id")),
			Type:     msg.Subject,
			Data:     msg.Data,
			Occurred: time.Now(),
		}

		evt, err := decodeEvent(envelope, &event.GraphCompletedEvent{})
		if err != nil {
			s.logger.Error("failed to decode event", "eventID", envelope.ID, "subject", envelope.Type, "error", err)
			return err
		}

		err = handler(ctx, evt.(*event.GraphCompletedEvent))
		duration := time.Since(start)

		if err != nil {
			s.logger.Error("handler failed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.GraphCompletedEvent).ScanID, "duration", duration, "error", err)
		} else {
			s.logger.Info("event processed", "eventID", envelope.ID, "subject", envelope.Type, "scanID", evt.(*event.GraphCompletedEvent).ScanID, "duration", duration)
		}

		return err
	}, opts...)
}

func encodeEvent(evt interface{}) ([]byte, error) {
	return json.Marshal(evt)
}

func decodeEvent(envelope *event.EventEnvelope, target interface{}) (interface{}, error) {
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		return nil, fmt.Errorf("failed to decode event: %w", err)
	}
	return target, nil
}

func encodeEnvelope(envelope event.EventEnvelope) ([]byte, error) {
	return json.Marshal(envelope)
}
