package manifestcontracts

type Config struct {
	DebugPrintOnly bool             `json:"debug_print_only,omitempty" yaml:"debug_print_only,omitempty"`
	OS             OSConfig         `json:"os,omitempty" yaml:"os,omitempty"`
	Kubernetes     KubernetesConfig `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`
}
