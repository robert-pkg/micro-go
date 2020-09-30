package log

// LogConfig .
type LogConfig struct {
	LogPath       string `yaml:"logPath"`
	Level         string `yaml:"level"`
	Encoding      string `yaml:"encoding"`
	OutputConsole bool   `yaml:"output_console"`
}
