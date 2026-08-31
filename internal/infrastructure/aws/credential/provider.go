package credential

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Provider interface {
	Load(ctx context.Context, region string) (aws.Config, error)
}

type DefaultProvider struct{}

func (p *DefaultProvider) Load(ctx context.Context, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	if region != "" {
		cfg.Region = region
	}

	if cfg.Region == "" {
		if r := regionFromEnv(); r != "" {
			cfg.Region = r
		}
	}

	if cfg.Region == "" {
		return aws.Config{}, fmt.Errorf("AWS region is not set")
	}

	return cfg, nil
}

type AssumeRoleConfig struct {
	RoleARN      string
	SessionName  string
	ExternalID   string
}

type AssumeRoleProvider struct {
	base   Provider
	config AssumeRoleConfig
}

func NewAssumeRoleProvider(base Provider, cfg AssumeRoleConfig) *AssumeRoleProvider {
	if cfg.SessionName == "" {
		cfg.SessionName = "weavelens-session"
	}
	return &AssumeRoleProvider{
		base:   base,
		config: cfg,
	}
}

func (p *AssumeRoleProvider) Load(ctx context.Context, region string) (aws.Config, error) {
	baseCfg, err := p.base.Load(ctx, region)
	if err != nil {
		return aws.Config{}, err
	}

	if p.config.RoleARN == "" {
		return baseCfg, nil
	}

	if !strings.HasPrefix(p.config.RoleARN, "arn:") {
		return aws.Config{}, ErrInvalidRoleARN
	}

	client := sts.NewFromConfig(baseCfg)
	input := &sts.AssumeRoleInput{
		RoleArn:         &p.config.RoleARN,
		RoleSessionName: &p.config.SessionName,
	}
	if p.config.ExternalID != "" {
		input.ExternalId = &p.config.ExternalID
	}

	output, err := client.AssumeRole(ctx, input)
	if err != nil {
		if isAccessDenied(err) {
			return aws.Config{}, fmt.Errorf("%w: %v", ErrAssumeRole, err)
		}
		return aws.Config{}, fmt.Errorf("%w: %v", ErrAssumeRole, err)
	}

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
		*output.Credentials.AccessKeyId,
		*output.Credentials.SecretAccessKey,
		*output.Credentials.SessionToken,
	))

	return aws.Config{
		Region:      baseCfg.Region,
		Credentials: creds,
	}, nil
}

func regionFromEnv() string {
	if r := os.Getenv("AWS_REGION"); r != "" {
		return r
	}
	if r := os.Getenv("AWS_DEFAULT_REGION"); r != "" {
		return r
	}
	return ""
}
