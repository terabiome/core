package natsjetstream

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/terabiome/core/pkg/logging"
	"github.com/terabiome/core/pkg/messaging"
)

type Config struct {
	URL string
}

type Client struct {
	cfg    Config
	conn   *nats.Conn
	js     jetstream.JetStream
	logger logging.Logger
	mu     sync.Mutex
}

// Publish implements [messaging.Publisher]
func (c *Client) Publish(
	ctx context.Context,
	destination messaging.Destination,
	data []byte,
) error {
	js, err := c.jetstream()
	if err != nil {
		return err
	}

	// ignore ACK for now
	_, err = js.Publish(ctx, destination.Subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish to %s: %w", destination.Subject, err)
	}
	return nil
}

// Consume implements [messaging.Consumer]
func (c *Client) Consume(
	ctx context.Context,
	source messaging.Source,
	processor messaging.Processor[messaging.Message],
) (*messaging.Subscription, error) {
	js, err := c.jetstream()
	if err != nil {
		return nil, err
	}

	stream, err := js.Stream(ctx, source.Stream)
	if err != nil {
		return nil, fmt.Errorf("failed to access stream %s: %w", source.Stream, err)
	}

	consumerConfig, options, err := buildConsumerConfig(source)
	if err != nil {
		return nil, fmt.Errorf("failed to build consumer config: %w", err)
	}

	consumer, err := ensureConsumer(ctx, stream, consumerConfig, options)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize consumer for %s: %w", source.Subject, err)
	}

	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		handleErr := processor.Handle(ctx, Message{Raw: msg})
		switch {
		case handleErr == nil:
			if ackErr := msg.Ack(); ackErr != nil && c.logger != nil {
				c.logger.ErrorContext(ctx, "failed to acknowledge message",
					"subject", msg.Subject(),
					"error", ackErr)
			}
		case messaging.IsTerminal(handleErr):
			if c.logger != nil {
				c.logger.WarnContext(ctx, "terminating unprocessable message",
					"subject", msg.Subject(),
					"error", handleErr)
			}
			if termErr := msg.Term(); termErr != nil && c.logger != nil {
				c.logger.ErrorContext(ctx, "failed to terminate message",
					"subject", msg.Subject(),
					"error", termErr)
			}
		default:
			if c.logger != nil {
				c.logger.ErrorContext(ctx, "message processing failed",
					"subject", msg.Subject(),
					"error", handleErr)
			}
			if nakErr := msg.Nak(); nakErr != nil && c.logger != nil {
				c.logger.ErrorContext(ctx, "failed to negatively acknowledge message",
					"subject", msg.Subject(),
					"error", nakErr)
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start consumption for %s: %w", source.Subject, err)
	}

	return messaging.NewSubscription(
		source,
		consumeCtx.Stop,
		consumeCtx.Drain,
		consumeCtx.Closed(),
	), nil
}

func ensureConsumer(
	ctx context.Context,
	stream jetstream.Stream,
	config *jetstream.ConsumerConfig,
	options SubscribeOptions,
) (jetstream.Consumer, error) {
	if options.Override {
		return stream.CreateOrUpdateConsumer(ctx, *config)
	}
	if options.Ephemeral {
		return stream.CreateConsumer(ctx, *config)
	}

	consumer, err := stream.Consumer(ctx, config.Durable)
	if errors.Is(err, jetstream.ErrConsumerNotFound) {
		return stream.CreateConsumer(ctx, *config)
	}
	return consumer, err
}

// New constructs a client. It does not connect.
func New(cfg Config, logger logging.Logger) *Client {
	return &Client{cfg: cfg, logger: logger}
}

// Connect dials NATS. It is safe to call more than once.
func (c *Client) Connect() error {
	if c == nil {
		return errors.New("jetstream client is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil && !c.conn.IsClosed() {
		return nil
	}

	url := c.cfg.URL
	if url == "" {
		url = nats.DefaultURL
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create Jetstream instance: %w", err)
	}

	c.conn = conn
	c.js = js
	return nil
}

func (c *Client) jetstream() (jetstream.JetStream, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.js, nil
}

// Close flushes in-flight publishes and closes the connection.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}
	return c.conn.Drain()
}
