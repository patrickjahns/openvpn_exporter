package config

// Server defines the general server configuration.
type Server struct {
	Addr string
	Path string
	Root string
}

// Logs defines the level for configuration
type Logs struct {
	Level string
}

// Config defines the general configuration object
type Config struct {
	Server          Server
	Logs            Logs
	StatusCollector StatusCollector
	ExportGoMetrics bool
}

// StatusCollector contains configuration for the OpenVPN status collector
type StatusCollector struct {
	ExportClientMetrics bool
	StatusFile          []string
}

// Load initializes a default configuration struct.
func Load() *Config {
	return &Config{}
}
