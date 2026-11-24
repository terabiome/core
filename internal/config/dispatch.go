package config

import (
	"log"
	"strconv"
	"time"
)

type DispatchConfig struct {
	SweepLimit       int
	SweepInterval    time.Duration
	SweepPassTimeout time.Duration
}

func NewDispatchConfig() DispatchConfig {
	return DispatchConfig{
		SweepLimit:       loadPositiveInt("DISPATCH_SWEEP_LIMIT", 32),
		SweepInterval:    loadPositiveDuration("DISPATCH_SWEEP_INTERVAL", 5*time.Second),
		SweepPassTimeout: loadPositiveDuration("DISPATCH_SWEEP_PASS_TIMEOUT", 2*time.Minute),
	}
}

func loadPositiveInt(key string, fallback int) int {
	raw := loadEnv[string](key, false)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		log.Fatalf("failed to load %s: failed to cast to int: %v\n", key, err)
	}
	if value <= 0 {
		log.Fatalf("failed to load %s: must be > 0\n", key)
	}
	return value
}

func loadPositiveDuration(key string, fallback time.Duration) time.Duration {
	raw := loadEnv[string](key, false)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		log.Fatalf("failed to load %s: failed to cast to duration: %v\n", key, err)
	}
	if value <= 0 {
		log.Fatalf("failed to load %s: must be > 0\n", key)
	}
	return value
}
