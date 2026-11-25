package messaging

import (
	"context"
	"time"
)

type Message interface {
	Data() []byte
	Ack() error
	Nak() error
	NakWithDelay(delay time.Duration) error
	InProgress() error
	Term() error
	Headers() map[string][]string
}

type Publisher interface {
	Publish(ctx context.Context, destination Destination, data []byte) error
}

type Consumer[M Message, S any] interface {
	Consume(ctx context.Context, source Source, processor Processor[M]) (*Subscription[S], error)
}

type Transport[M Message, S any] interface {
	Publisher
	Consumer[M, S]
}

type Processor[M Message] interface {
	Handle(ctx context.Context, message M) error
}
