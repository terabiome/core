package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/terabiome/terabiome/internal/messaging"
	"github.com/terabiome/terabiome/pkg/logging"
)

type NATSJetstream struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// Produce implements [messaging.Producer]
func (c *NATSJetstream) Produce(
	ctx context.Context,
	destination messaging.Destination,
	data []byte,
) error {
	_, err := c.js.Publish(ctx, destination.Topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish to stream %s: %w", destination.Topic, err)
	}
	return nil
}

// Consume implements [messaging.Consumer]
func (c *NATSJetstream) Consume(
	ctx context.Context,
	source messaging.Source,
	processor messaging.Processor[Message],
) (*messaging.Subscription[jetstream.ConsumeContext], error) {
	stream, err := c.js.Stream(ctx, source.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to access stream %s: %w", source.Topic, err)
	}

	consumerConfig, err := buildJetstreamConsumerConfig(source.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to build consumer config: %w", err)
	}

	var consumer jetstream.Consumer

	if override, ok := source.Parameters["override_consumer"].(bool); ok && override {
		consumer, err = stream.CreateOrUpdateConsumer(ctx, *consumerConfig)
	} else {
		consumer, err = stream.Consumer(ctx, consumerConfig.Durable)
		if err == jetstream.ErrConsumerNotFound {
			consumer, err = stream.CreateConsumer(ctx, *consumerConfig)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize consumer for %s: %w", source.Topic, err)
	}

	// Start the consumption loop
	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		if handleErr := processor.Handle(Message{msg}); handleErr != nil {
			logging.GetGlobalSlogger().ErrorContext(ctx, "message processing failed",
				"subject", msg.Subject(),
				"error", handleErr)
			msg.Nak() // Return to stream for retry
			return
		}
		msg.Ack() // Confirm success
	})

	if err != nil {
		return nil, fmt.Errorf("failed to start consumption for %s: %w", source.Topic, err)
	}

	return &messaging.Subscription[jetstream.ConsumeContext]{
		Source: source,
		Raw:    consumeCtx,
	}, nil
}

// Constructor
func NewNATSJetstream() (*NATSJetstream, error) {
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jetstream instance: %w", err)
	}

	return &NATSJetstream{conn: conn, js: js}, nil
}
