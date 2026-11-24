package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Logger   LoggerConfig
	Database DatabaseConfig
	NATS     NATSConfig
	Dispatch DispatchConfig
}

var cfg *Config

func New() *Config {
	content, err := os.ReadFile(".env")
	if err != nil {
		log.Println("fallback to ENV read")
	} else {
		for line := range strings.Lines(string(content)) {
			line = strings.TrimFunc(line, func(r rune) bool {
				if r == '\n' || r == '\r' || r == ' ' {
					return true
				}
				return false
			})
			sepIdx := strings.Index(line, "=")
			if sepIdx <= 0 {
				log.Fatalln("malformed k-v, crashing")
			}
			k, v := line[:sepIdx], line[sepIdx+1:]
			if err := os.Setenv(k, v); err != nil {
				log.Fatalln("failed to set ENV, crashing")
			}
		}
	}

	cfg := Config{
		Logger:   NewLoggerConfig(),
		Database: NewDatabaseConfig(),
		NATS:     NewNATSConfig(),
		Dispatch: NewDispatchConfig(),
	}
	return &cfg
}

func Get() *Config {
	if cfg == nil {
		cfg = New()
	}

	return cfg
}

// can log fatal here
func loadEnv[T any](key string, required bool) T {
	value, exist := os.LookupEnv(key)
	if !exist && required {
		log.Fatalf("failed to load %s: not exist\n", key)
	}

	var anchor T
	switch anchorPtr := any(&anchor).(type) {
	case *string:
		*anchorPtr = value
	case *int:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("failed to load %s: failed to cast to int: %v\n", key, err)
		}
		*anchorPtr = intVal
	case *bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			log.Fatalf("failed to load %s: failed to cast to bool: %v\n", key, err)
		}
		*anchorPtr = boolVal
	}

	return anchor
}
