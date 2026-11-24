package manifestconfig

import "go.yaml.in/yaml/v4"

type KubernetesDistribution string

const (
	KubernetesDistributionK3s KubernetesDistribution = "k3s"
)

type KubernetesNodeRole string

const (
	KubernetesNodeRoleControlPlane KubernetesNodeRole = "control-plane"
	KubernetesNodeRoleWorker       KubernetesNodeRole = "worker"
)

type KubernetesConfig struct {
	Distribution   KubernetesDistribution `json:"distribution" yaml:"distribution"`
	NodeRole       KubernetesNodeRole     `json:"node_role" yaml:"node_role"`
	Version        string                 `json:"version" yaml:"version"` // empty -> fallback to "latest"
	Configurations yaml.Node              `json:"configurations" yaml:"configurations"`
}

type K3sConfigurations struct {
	ServerURL string `json:"server_url" yaml:"server_url"`
	Token     string `json:"token" yaml:"token"`
}

type RKE2Configurations struct {
}
