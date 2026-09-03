package credential

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
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
	ensureHomeDir()

	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
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

// ensureHomeDir ensures that the HOME, USERPROFILE, and AWS config file
// path environment variables are set so that the AWS SDK can resolve
// user-level configuration files such as ~/.aws/credentials and
// ~/.aws/config.
//
// This is primarily needed on Windows where HOME is not set by default —
// only USERPROFILE is. The AWS SDK v2 uses os.UserHomeDir() which checks
// USERPROFILE, but the SDK's DefaultSharedConfigFiles variable is evaluated
// at package init time. If USERPROFILE is unset in certain execution contexts
// (e.g. scheduled tasks, services), the resolved paths will be wrong.
// This function resolves the home directory from HOME → USERPROFILE →
// user.Current() and sets all variables explicitly.
func ensureHomeDir() {
	home := os.Getenv("HOME")
	homeWasSet := home != ""

	if !homeWasSet {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		if u, err := user.Current(); err == nil {
			home = u.HomeDir
		}
	}

	if home == "" {
		return
	}

	os.Setenv("HOME", home)
	if os.Getenv("USERPROFILE") == "" {
		os.Setenv("USERPROFILE", home)
	}

	if !homeWasSet {
		awsDir := filepath.Join(home, ".aws")
		if os.Getenv("AWS_CONFIG_FILE") == "" {
			os.Setenv("AWS_CONFIG_FILE", filepath.Join(awsDir, "config"))
		}
		if os.Getenv("AWS_SHARED_CREDENTIALS_FILE") == "" {
			os.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(awsDir, "credentials"))
		}
	}
}

type AssumeRoleConfig struct {
	RoleARN     string
	SessionName string
	ExternalID  string
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
