package natsjetstream

import (
	"github.com/nats-io/nats.go/jetstream"
)

type Message struct {
	Raw jetstream.Msg
}

// Ack implements [messaging.Message].
func (m Message) Ack() error {
	return m.Raw.Ack()
}

// Data implements [messaging.Message].
func (m Message) Data() []byte {
	return m.Raw.Data()
}

// Headers implements [messaging.Message].
func (m Message) Headers() map[string][]string {
	return m.Raw.Headers()
}

// Nak implements [messaging.Message].
func (m Message) Nak() error {
	return m.Raw.Nak()
}
