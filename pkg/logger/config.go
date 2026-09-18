package logger

// Config holds logger configuration.
//
// ENV variables:
//
//	LOGGER_LEVEL  — log level: debug | info | warn | error (default: info)
//	LOGGER_FOLDER — path to folder for log files (default: ./logs)
type Config struct {
	Level  string
	Folder string
}

// DefaultConfig returns sensible defaults for local development.
func DefaultConfig() Config {
	return Config{
		Level:  "debug",
		Folder: "./logs",
	}
}
