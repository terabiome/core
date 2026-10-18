package k3s

const (
	defaultK3sInstallBaseURL = "https://github.com/k3s-io/k3s/releases/download"
	defaultK3sBinaryPath     = "/usr/local/bin/k3s"
)

type installBinaryParameters struct {
	Version string `mapstructure:"version"`
}
