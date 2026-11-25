package k3s

const (
	defaultK3sInstallBaseURL = "https://github.com/k3s-io/k3s/releases/download"
	defaultK3sBinaryPath     = "/usr/local/bin/k3s"
	defaultK3sConfigPath     = "/etc/rancher/k3s/config.yaml"
)

type installBinaryParameters struct {
	Version string `mapstructure:"version"`
}

type k3sConfigEntry struct {
	Path   string         `mapstructure:"path"`
	Config map[string]any `mapstructure:"config"`
}
