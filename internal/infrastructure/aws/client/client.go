package client

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func LoadConfig(ctx context.Context, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	if region != "" {
		cfg.Region = region
	}

	if cfg.Region == "" {
		if r := os.Getenv("AWS_REGION"); r != "" {
			cfg.Region = r
		} else if r := os.Getenv("AWS_DEFAULT_REGION"); r != "" {
			cfg.Region = r
		}
	}

	if cfg.Region == "" {
		return aws.Config{}, fmt.Errorf("AWS region is not set")
	}

	return cfg, nil
}
