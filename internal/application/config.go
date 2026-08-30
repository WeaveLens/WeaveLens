package application

import (
	"os"
)

type Config struct {
	ServerPort string
	Env        string
	LogLevel   string
}

func Load() (*Config, error) {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		ServerPort: port,
		Env:        env,
		LogLevel:   logLevel,
	}, nil
}
