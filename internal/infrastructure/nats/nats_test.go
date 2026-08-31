package nats_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elip/WeaveLens/internal/domain/event"
	"github.com/elip/WeaveLens/internal/infrastructure/nats"
	gnats "github.com/nats-io/nats.go"
)

func startNATS(t *testing.T) *nats.Client {
	t.Helper()
	cfg := nats.NewConfig()
	cfg.URL = "nats://localhost:4222"
	cfg.DurablePrefix = "weavelens-test-" + t.Name()

	rawConn, err := gnats.Connect(cfg.URL, gnats.Timeout(5*time.Second))
	if err == nil {
		js, jsErr := rawConn.JetStream()
		if jsErr == nil {
			js.DeleteStream("weavelens")
		}
		rawConn.Close()
	}

	client, err := nats.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

func TestPublishScanStarted(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)

	evt := &event.ScanStartedEvent{
		ScanID: "scan-123",
		Region: "us-east-1",
	}

	err := publisher.PublishScanStarted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanStarted() error = %v", err)
	}
}

func TestPublishScanCompleted(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)

	evt := &event.ScanCompletedEvent{
		ScanID:        "scan-123",
		ResourceCount: 10,
	}

	err := publisher.PublishScanCompleted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanCompleted() error = %v", err)
	}
}

func TestPublishScanFailed(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)

	evt := &event.ScanFailedEvent{
		ScanID: "scan-123",
		Error:  "connection timeout",
	}

	err := publisher.PublishScanFailed(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanFailed() error = %v", err)
	}
}

func TestPublishResourceDiscovered(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)

	evt := &event.ResourceDiscoveredEvent{
		ScanID: "scan-123",
		Resource: event.Resource{
			ID:       "res-1",
			Name:     "test",
			Type:     "EC2",
			Category: "compute",
		},
	}

	err := publisher.PublishResourceDiscovered(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishResourceDiscovered() error = %v", err)
	}
}

func TestPublishRelationshipDiscovered(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)

	evt := &event.RelationshipDiscoveredEvent{
		ScanID: "scan-123",
		Relationship: event.Relationship{
			ID:       "rel-1",
			SourceID: "vpc-1",
			TargetID: "subnet-1",
			Type:     "contains",
		},
	}

	err := publisher.PublishRelationshipDiscovered(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishRelationshipDiscovered() error = %v", err)
	}
}

func TestPublishGraphCompleted(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)

	evt := &event.GraphCompletedEvent{
		ScanID:    "scan-123",
		NodeCount: 10,
		EdgeCount: 5,
	}

	err := publisher.PublishGraphCompleted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishGraphCompleted() error = %v", err)
	}
}

func TestSubscribeScanStarted(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)
	subscriber := nats.NewJetStreamSubscriber(client)

	received := make(chan *event.ScanStartedEvent, 1)
	handler := func(ctx context.Context, evt *event.ScanStartedEvent) error {
		received <- evt
		return nil
	}

	_, err := subscriber.SubscribeScanStarted(context.Background(), handler)
	if err != nil {
		t.Fatalf("SubscribeScanStarted() error = %v", err)
	}

	evt := &event.ScanStartedEvent{
		ScanID: "scan-123",
		Region: "us-east-1",
	}

	err = publisher.PublishScanStarted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanStarted() error = %v", err)
	}

	select {
	case msg := <-received:
		if msg.ScanID != "scan-123" {
			t.Errorf("SubscribeScanStarted() ScanID = %v, want scan-123", msg.ScanID)
		}
	case <-time.After(5 * time.Second):
		t.Error("SubscribeScanStarted() timeout waiting for message")
	}
}

func TestIdempotentProcessing(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)
	subscriber := nats.NewJetStreamSubscriber(client)

	var count int64
	handler := func(ctx context.Context, evt *event.ScanStartedEvent) error {
		atomic.AddInt64(&count, 1)
		return nil
	}

	_, err := subscriber.SubscribeScanStarted(context.Background(), handler)
	if err != nil {
		t.Fatalf("SubscribeScanStarted() error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	evt := &event.ScanStartedEvent{
		ScanID: "scan-123",
		Region: "us-east-1",
	}

	err = publisher.PublishScanStarted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanStarted() error = %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt64(&count) != 1 {
		t.Errorf("IdempotentProcessing() count = %v, want 1", atomic.LoadInt64(&count))
	}
}

func TestRetryOnHandlerFailure(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)
	subscriber := nats.NewJetStreamSubscriber(client)

	var count int64
	handler := func(ctx context.Context, evt *event.ScanStartedEvent) error {
		newCount := atomic.AddInt64(&count, 1)
		if newCount < 3 {
			return fmt.Errorf("transient error %d", newCount)
		}
		return nil
	}

	_, err := subscriber.SubscribeScanStarted(context.Background(), handler,
		nats.WithMaxDeliver(5),
		nats.WithAckWait(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("SubscribeScanStarted() error = %v", err)
	}

	evt := &event.ScanStartedEvent{
		ScanID: "scan-456",
		Region: "us-east-1",
	}

	err = publisher.PublishScanStarted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanStarted() error = %v", err)
	}

	time.Sleep(3 * time.Second)

	if atomic.LoadInt64(&count) != 3 {
		t.Errorf("RetryOnHandlerFailure() count = %v, want 3", atomic.LoadInt64(&count))
	}
}

func TestGracefulShutdown(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)
	subscriber := nats.NewJetStreamSubscriber(client)

	received := make(chan *event.ScanStartedEvent, 1)
	handler := func(ctx context.Context, evt *event.ScanStartedEvent) error {
		received <- evt
		return nil
	}

	sub, err := subscriber.SubscribeScanStarted(context.Background(), handler)
	if err != nil {
		t.Fatalf("SubscribeScanStarted() error = %v", err)
	}

	evt := &event.ScanStartedEvent{
		ScanID: "scan-789",
		Region: "us-west-2",
	}

	err = publisher.PublishScanStarted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanStarted() error = %v", err)
	}

	select {
	case msg := <-received:
		if msg.ScanID != "scan-789" {
			t.Errorf("GracefulShutdown() ScanID = %v, want scan-789", msg.ScanID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("GracefulShutdown() timeout waiting for message")
	}

	done := make(chan struct{})
	go func() {
		sub.Drain()
		close(done)
	}()

	select {
	case <-done:
		// drained successfully
	case <-time.After(5 * time.Second):
		t.Fatal("GracefulShutdown() timeout waiting for drain")
	}
}

func TestConsumerRestart(t *testing.T) {
	client := startNATS(t)
	publisher := nats.NewJetStreamPublisher(client)
	subscriber := nats.NewJetStreamSubscriber(client)

	received := make(chan *event.ScanStartedEvent, 1)
	handler := func(ctx context.Context, evt *event.ScanStartedEvent) error {
		received <- evt
		return nil
	}

	_, err := subscriber.SubscribeScanStarted(context.Background(), handler, nats.WithDurable("restart-consumer"))
	if err != nil {
		t.Fatalf("SubscribeScanStarted() error = %v", err)
	}

	evt := &event.ScanStartedEvent{
		ScanID: "scan-restart",
		Region: "eu-west-1",
	}

	err = publisher.PublishScanStarted(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishScanStarted() error = %v", err)
	}

	select {
	case msg := <-received:
		if msg.ScanID != "scan-restart" {
			t.Errorf("ConsumerRestart() ScanID = %v, want scan-restart", msg.ScanID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ConsumerRestart() timeout waiting for message")
	}
}
