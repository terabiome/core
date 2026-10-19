package nats

import (
	"github.com/nats-io/nats.go/jetstream"
)

type Message struct {
	msg jetstream.Msg
}

// Acknowledge implements [messaging.Message].
func (m Message) Acknowledge() error {
	return m.msg.Ack()
}

// Data implements [messaging.Message].
func (m Message) Data() []byte {
	return m.msg.Data()
}

// Headers implements [messaging.Message].
func (m Message) Headers() map[string][]string {
	return m.msg.Headers()
}

// Reject implements [messaging.Message].
func (m Message) Reject() error {
	return m.msg.Nak()
}
