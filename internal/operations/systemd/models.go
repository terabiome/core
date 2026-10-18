package systemd

type configureSystemdUnitParameters struct {
	Name    string `mapstructure:"name"`    // e.g., "k3s"
	Content string `mapstructure:"content"` // The full .service file content
}
