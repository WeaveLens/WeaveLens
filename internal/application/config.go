package application

import (
	"os"
)

type Config struct {
	ServerPort        string
	Env               string
	LogLevel          string
	NATSURL           string
	AWSRegion         string
	AWSRoleARN        string
	AWSRoleSessionName string
	AWSExternalID     string
	APIKey            string
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

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = os.Getenv("AWS_DEFAULT_REGION")
	}

	return &Config{
		ServerPort:        port,
		Env:               env,
		LogLevel:          logLevel,
		NATSURL:           natsURL,
		AWSRegion:         awsRegion,
		AWSRoleARN:        os.Getenv("AWS_ROLE_ARN"),
		AWSRoleSessionName: os.Getenv("AWS_ROLE_SESSION_NAME"),
		AWSExternalID:     os.Getenv("AWS_EXTERNAL_ID"),
		APIKey:            os.Getenv("API_KEY"),
	}, nil
}
