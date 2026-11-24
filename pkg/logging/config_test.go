package logging_test

import (
	"testing"

	"github.com/terabiome/core/pkg/logging"
)

func TestPublicLoggingConfig(t *testing.T) {
	t.Parallel()

	slogger := logging.NewSlogger(
		logging.Config{Level: "debug", Format: "json"},
		map[string]any{"scope": "test"},
	)
	if slogger == nil {
		t.Fatal("NewSlogger() = nil")
	}
}
