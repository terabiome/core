package kernel

type action string

const (
	actionAdd    action = "add"
	actionRemove action = "remove"
)

type kernelModuleConfEntry struct {
	Path   string `mapstructure:"path"`
	Value  string `mapstructure:"value"`
	Action action `mapstructure:"action"`
}

type sysctlConfEntry struct {
	Path  string `mapstructure:"path"`
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
}
