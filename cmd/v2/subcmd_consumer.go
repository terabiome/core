package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/terabiome/core/pkg/consumer"
	"github.com/terabiome/core/pkg/logging"
	"github.com/terabiome/core/pkg/messaging/natsjetstream"
)

const consumerSubcommand subcommand = "consumer"

type consumerCLIInput struct {
	NATSURL     string        `validate:"required"`
	Stream      string        `validate:"required"`
	Subjects    stringList    `validate:"required,min=1"`
	Concurrency int           `validate:"min=0"`
	AckWait     time.Duration `validate:"min=0"`
	MaxDeliver  int           `validate:"min=0"`
	Durable     string
	Override    bool
}

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	*s = append(*s, value)
	return nil
}

func processConsumerSubcommand(ctx context.Context, input consumerCLIInput) error {
	subscriptions := make([]consumer.SubscriptionConfig, 0, len(input.Subjects))
	for i, subject := range input.Subjects {
		options := natsjetstream.SubscribeOptions{
			AckWait:    input.AckWait,
			MaxDeliver: input.MaxDeliver,
			Override:   input.Override,
		}
		if input.Durable != "" {
			options.Durable = input.Durable
			if len(input.Subjects) > 1 {
				options.Durable = fmt.Sprintf("%s_%d", input.Durable, i)
			}
		}

		subscriptions = append(subscriptions, consumer.SubscriptionConfig{
			Stream:  input.Stream,
			Subject: subject,
			Options: options,
		})
	}

	worker, err := consumer.NewWorker(
		consumer.Config{
			Jetstream:     natsjetstream.Config{URL: input.NATSURL},
			Subscriptions: subscriptions,
			Concurrency:   input.Concurrency,
		},
		&(consumer.Dependencies{}),
		logging.GetGlobalSlogger(),
	)
	if err != nil {
		return err
	}

	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
