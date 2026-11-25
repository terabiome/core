package kernel

type kernelModuleConfEntry struct {
	Path   string `mapstructure:"path"`
	Value  string `mapstructure:"value"`
	Action string `mapstructure:"action"` // "add" or "remove"
}

type sysctlConfEntry struct {
	Path  string `mapstructure:"path"`
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
}
