package nats

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elip/WeaveLens/internal/domain/event"
	"github.com/nats-io/nats.go"
)

type Client struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config *Config
}

type Config struct {
	URL            string
	StreamName     string
	MaxAge         time.Duration
	MaxBytes       int64
	MaxMsgs        int64
	DurablePrefix  string
	AckWait        time.Duration
	MaxDeliver     int
	Backoff        []time.Duration
}

func NewConfig() *Config {
	return &Config{
		URL:           "nats://localhost:4222",
		StreamName:    "weavelens",
		MaxAge:        24 * time.Hour,
		MaxMsgs:       1000000,
		DurablePrefix: "weavelens",
		AckWait:       30 * time.Second,
		MaxDeliver:    1,
		Backoff:       []time.Duration{time.Second, 5 * time.Second, 15 * time.Second},
	}
}

func Connect(ctx context.Context, cfg *Config) (*Client, error) {
	conn, err := nats.Connect(cfg.URL,
		nats.MaxReconnects(10),
		nats.ReconnectWait(1*time.Second),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	client := &Client{
		conn:   conn,
		js:     js,
		config: cfg,
	}

	if err := client.ensureStream(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ensure stream: %w", err)
	}

	return client, nil
}

func (c *Client) ensureStream(ctx context.Context) error {
	subjects := []string{
		"weavelens.scan.started.v1",
		"weavelens.scan.completed.v1",
		"weavelens.scan.failed.v1",
		"weavelens.resource.discovered.v1",
		"weavelens.relationship.discovered.v1",
		"weavelens.graph.completed.v1",
	}

	streamConfig := &nats.StreamConfig{
		Name:     c.config.StreamName,
		Subjects: subjects,
		MaxAge:   c.config.MaxAge,
		MaxBytes: c.config.MaxBytes,
		MaxMsgs:  c.config.MaxMsgs,
		Storage:  nats.MemoryStorage,
	}

	_, err := c.js.AddStream(streamConfig)
	if err != nil {
		if err == nats.ErrStreamNameAlreadyInUse {
			return nil
		}
		if strings.Contains(err.Error(), "subjects overlap") {
			return nil
		}
		return err
	}

	return nil
}

func (c *Client) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish to %s: %w", subject, err)
	}
	return nil
}

func (c *Client) Subscribe(subject string, handler func(*nats.Msg) error, opts ...SubscriberOption) (*Subscription, error) {
	options := &SubscriberOptions{
		Durable:     strings.ReplaceAll(c.config.DurablePrefix+"."+strings.ReplaceAll(subject, ".", "-"), ".", "-"),
		AckWait:     c.config.AckWait,
		MaxDeliver:  c.config.MaxDeliver,
		Backoff:     c.config.Backoff,
	}

	for _, opt := range opts {
		opt(options)
	}

	sub, err := c.js.SubscribeSync(
		subject,
		nats.Durable(options.Durable),
		nats.AckWait(options.AckWait),
		nats.MaxDeliver(options.MaxDeliver),
		nats.MaxAckPending(1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	subscription := &Subscription{
		subject: subject,
		handler: handler,
		sub:     sub,
		done:    make(chan struct{}),
	}

	go subscription.run()

	return subscription, nil
}

func (c *Client) Close() error {
	if c.conn != nil && !c.conn.IsClosed() {
		c.conn.Close()
	}
	return nil
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

type MessageHandler func(ctx context.Context, msg *event.EventEnvelope) error

type SubscriberOptions struct {
	Durable     string
	AckWait     time.Duration
	MaxDeliver  int
	Backoff     []time.Duration
}

type SubscriberOption func(*SubscriberOptions)

func WithDurable(durable string) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.Durable = durable
	}
}

func WithAckWait(ackWait time.Duration) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.AckWait = ackWait
	}
}

func WithMaxDeliver(maxDeliver int) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.MaxDeliver = maxDeliver
	}
}

func WithBackoff(backoff []time.Duration) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.Backoff = backoff
	}
}

type Subscription struct {
	subject string
	handler func(*nats.Msg) error
	sub     *nats.Subscription
	done    chan struct{}
}

func (s *Subscription) run() {
	defer close(s.done)

	for {
		select {
		case <-s.done:
			return
		default:
			msg, err := s.sub.NextMsg(1 * time.Minute)
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				if err == nats.ErrConnectionClosed {
					return
				}
				continue
			}

			if err := s.handler(msg); err != nil {
				msg.Nak()
				continue
			}

			msg.Ack()
		}
	}
}

func (s *Subscription) Drain() error {
	return s.sub.Drain()
}

func (s *Subscription) Unsubscribe() error {
	return s.sub.Unsubscribe()
}

func (s *Subscription) Done() <-chan struct{} {
	return s.done
}
