package logger

// LOGGER_LEVEL  — log level: debug | info | warn | error (default: info)
// LOGGER_FOLDER — path to folder for log files (default: ./logs)
type Config struct {
	Level  string
	Folder string
}

func DefaultConfig() Config {
	return Config{
		Level:  "debug",
		Folder: "./logs",
	}
}
