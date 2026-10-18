package shell

type cmdParameters struct {
	Executable             string   `mapstructure:"-"`
	Arguments              []string `mapstructure:"-"`
	Command                string   `mapstructure:"command"`
	CommandSegments        []string `mapstructure:"command_segments"`
	LogCommandInStdStreams bool     `mapstructure:"log_command_in_std_streams"`
}
