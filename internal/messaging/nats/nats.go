package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/terabiome/terabiome/internal/messaging"
)

type NATSJetstream struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

func (c *NATSJetstream) Produce(
	ctx context.Context,
	destination messaging.Destination,
	data []byte,
) error {
	opts := []jetstream.PublishOpt{}

	// think of something to do with *jetstream.PubAck
	_, err := c.js.Publish(ctx, destination.Topic, data, opts...)

	if err != nil {
		return fmt.Errorf("failed to publish to stream %s: %w", destination.Topic, err)
	}
	return nil
}

func (c *NATSJetstream) Consume(
	ctx context.Context,
	source messaging.Source,
	processor messaging.Processor[Message],
) (*messaging.Subscription[jetstream.ConsumeContext], error) {
	stream, err := c.js.Stream(ctx, source.Topic)
	if err == jetstream.ErrStreamNotFound {
		return nil, fmt.Errorf("connected to non-existent stream %s", source.Topic)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stream: %w", err)
	}

	consumerConfig, err := buildJetstreamConsumerConfig(source.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to build jetstream.ConsumerConfig: %w", err)
	}

	var consumer jetstream.Consumer
	if source.Parameters["override_consumer"].(bool) {
		consumer, err = stream.CreateOrUpdateConsumer(ctx, *consumerConfig)
	} else {
		consumer, err = stream.CreateConsumer(ctx, *consumerConfig)
		if err == jetstream.ErrConsumerExists {
			return nil, fmt.Errorf("consumer already exists")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer for stream %s: %w", source.Topic, err)
	}

	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		if handleErr := processor.Handle(Message{msg}); handleErr != nil {
			// do something
		}
		// do something else
	})

	if err != nil {
		return nil, fmt.Errorf("failed to consume stream %v: %w", source.Topic, err)
	}

	return &messaging.Subscription[jetstream.ConsumeContext]{
		Source: source,
		Raw:    consumeCtx,
	}, nil
}

// NewNATSJetstream is constructor
func NewNATSJetstream() (messaging.Bridge[Message, jetstream.ConsumeContext], error) {
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jetstream instance: %w", err)
	}

	client := NATSJetstream{
		conn: conn,
		js:   js,
	}

	return &client, nil
}
