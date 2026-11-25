package natsjetstream

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/terabiome/core/pkg/messaging"
)

type SubscribeOptions struct {
	Name          string
	Description   string
	Durable       string
	AckWait       time.Duration
	MaxDeliver    int
	MaxAckPending int
	BackOff       []time.Duration
	Override      bool // replaces an existing consumer with the same durable name.
	Ephemeral     bool // creates an unnamed consumer that is removed once disconnected.
}

type legacyOptions struct {
	Name        string `mapstructure:"consumer_name"`
	Description string `mapstructure:"consumer_description"`
}

func buildConsumerConfig(source messaging.Source) (*jetstream.ConsumerConfig, SubscribeOptions, error) {
	var options SubscribeOptions

	switch typed := source.Options.(type) {
	case nil:
	case SubscribeOptions:
		options = typed
	case *SubscribeOptions:
		if typed != nil {
			options = *typed
		}
	case map[string]any:
		var legacy legacyOptions
		if err := mapstructure.Decode(typed, &legacy); err != nil {
			return nil, SubscribeOptions{}, fmt.Errorf("failed to decode consumer options: %w", err)
		}
		options.Name = legacy.Name
		options.Description = legacy.Description
	default:
		return nil, SubscribeOptions{}, fmt.Errorf("unsupported consumer options type %T", source.Options)
	}

	if options.Durable == "" && !options.Ephemeral {
		if options.Name != "" {
			options.Durable = options.Name
		} else {
			options.Durable = defaultDurableName(source.Stream, source.Subject)
		}
	}

	if options.Override && options.Ephemeral {
		return nil, SubscribeOptions{}, errors.New("override and ephemeral are mutually exclusive")
	}

	config := &jetstream.ConsumerConfig{
		Name:          options.Name,
		Description:   options.Description,
		Durable:       options.Durable,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       options.AckWait,
		MaxDeliver:    options.MaxDeliver,
		MaxAckPending: options.MaxAckPending,
		BackOff:       options.BackOff,
		FilterSubject: source.Subject,
	}

	return config, options, nil
}

func defaultDurableName(stream, subject string) string {
	replacer := strings.NewReplacer(".", "_", "*", "_", ">", "_", " ", "_")
	name := replacer.Replace(fmt.Sprintf("%s_%s", stream, subject))
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}
