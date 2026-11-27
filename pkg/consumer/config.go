package consumer

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/terabiome/core/pkg/journalers"
	"github.com/terabiome/core/pkg/messaging/natsjetstream"
	"github.com/terabiome/core/pkg/operations"
)

const defaultShutdownTimeout = 30 * time.Second

type SubscriptionConfig struct {
	Stream  string
	Subject string
	Options natsjetstream.SubscribeOptions
}

type JournalConfig struct {
	MaxRecords    int
	FlushInterval time.Duration
	BatchFn       journalers.BatchHandler
}

type Config struct {
	Jetstream       natsjetstream.Config
	Subscriptions   []SubscriptionConfig
	Node            operations.NodeInfo
	Concurrency     int
	Journal         JournalConfig
	ShutdownTimeout time.Duration
}

func (c *Config) normalize() error {
	if len(c.Subscriptions) == 0 {
		return errors.New("at least one subscription is required")
	}
	for i, subscription := range c.Subscriptions {
		if subscription.Stream == "" {
			return fmt.Errorf("subscription %d: stream cannot be empty", i)
		}
		if subscription.Subject == "" {
			return fmt.Errorf("subscription %d: subject cannot be empty", i)
		}
	}

	if c.Node.MachineID == "" {
		if hostname, err := os.Hostname(); err == nil {
			c.Node.MachineID = hostname
		}
	}
	if c.Node.Arch == "" {
		c.Node.Arch = runtime.GOARCH
	}
	if c.Node.OS == "" {
		c.Node.OS = runtime.GOOS
	}

	if c.Concurrency <= 0 {
		c.Concurrency = runtime.NumCPU()
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = defaultShutdownTimeout
	}
	return nil
}
