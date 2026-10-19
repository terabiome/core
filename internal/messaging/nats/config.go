package nats

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/nats-io/nats.go/jetstream"
)

type jetstreamConsumerInternalConfig struct {
	Name        string `mapstructure:"consumer_name"`
	Description string `mapstructure:"consumer_description"`
	// more to come
}

func buildJetstreamConsumerConfig(parameters map[string]any) (*jetstream.ConsumerConfig, error) {
	var internalRepr jetstreamConsumerInternalConfig
	if err := mapstructure.Decode(parameters, &internalRepr); err != nil {
		return nil, fmt.Errorf("failed to convert parameter map into %T: %w", internalRepr, err)
	}

	return &jetstream.ConsumerConfig{
		Name:        internalRepr.Name,
		Description: internalRepr.Description,
		// more to come
	}, nil
}
