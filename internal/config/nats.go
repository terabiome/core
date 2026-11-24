package config

type NATSConfig struct {
	URL string
}

func NewNATSConfig() NATSConfig {
	return NATSConfig{
		URL: loadEnv[string]("NATS_URL", false),
	}
}
