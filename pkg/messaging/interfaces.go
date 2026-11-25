package messaging

import (
	"context"
)

type Message interface {
	Data() []byte
	Acknowledge() error
	Reject() error
	Headers() map[string][]string
}

type Producer interface {
	Produce(ctx context.Context, destination Destination, data []byte) error
}

type Consumer[M Message, S any] interface {
	Consume(ctx context.Context, source Source, processor Processor[M]) (*Subscription[S], error)
}

type Bridge[M Message, S any] interface {
	Producer
	Consumer[M, S]
}

type Processor[M Message] interface {
	Handle(input M) error
}
