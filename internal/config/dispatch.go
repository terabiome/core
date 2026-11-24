package config

import (
	"log"
	"time"
)

type DispatchConfig struct {
	SweepPassTimeout time.Duration
}

func NewDispatchConfig() DispatchConfig {
	const defaultTimeout = 2 * time.Minute

	raw := loadEnv[string]("DISPATCH_SWEEP_PASS_TIMEOUT", false)
	if raw == "" {
		return DispatchConfig{SweepPassTimeout: defaultTimeout}
	}

	timeout, err := time.ParseDuration(raw)
	if err != nil {
		log.Fatalf("failed to load DISPATCH_SWEEP_PASS_TIMEOUT: failed to cast to duration: %v\n", err)
	}
	if timeout <= 0 {
		log.Fatalf("failed to load DISPATCH_SWEEP_PASS_TIMEOUT: must be > 0\n")
	}

	return DispatchConfig{SweepPassTimeout: timeout}
}
