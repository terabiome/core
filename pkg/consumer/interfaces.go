package consumer

import (
	"io"

	"github.com/terabiome/core/pkg/messaging"
	"github.com/terabiome/core/pkg/operations"
)

type Registry interface {
	Get(name string) (operations.Operation, error)
}

type MessagingConsumer interface {
	messaging.Consumer
	io.Closer
}

type Journaler interface {
	io.Writer
	io.Closer
}

type Handler = messaging.Processor[messaging.Message]
