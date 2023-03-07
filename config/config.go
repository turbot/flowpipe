package config

// Config is the configuration for the application.
type Config struct {
	// Directory where log files will be written.
	LogDir string `json:"log_dir,omitempty"`
}

type ConfigOption func(*Config)

// NewConfig creates a new Config instance with the given options.
func NewConfig(opts ...ConfigOption) *Config {
	const (
		defaultLogDir = "logs"
	)

	c := &Config{
		LogDir: defaultLogDir,
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *Config as the argument
		opt(c)
	}

	// return the modified config instance
	return c
}

// WithLogDir returns a ConfigOption that sets the log directory.
func WithLogDir(logDir string) ConfigOption {
	return func(c *Config) {
		c.LogDir = logDir
	}
}
