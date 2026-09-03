package credential

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Provider interface {
	Load(ctx context.Context, region string) (aws.Config, error)
}

type DefaultProvider struct {
	mu      sync.RWMutex
	profile string
	logger  *slog.Logger
}

func NewDefaultProvider(logger *slog.Logger) *DefaultProvider {
	return &DefaultProvider{logger: logger}
}

func (p *DefaultProvider) Profile() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.profile
}

func (p *DefaultProvider) setProfile(profile string) {
	p.mu.Lock()
	p.profile = profile
	p.mu.Unlock()
}

func (p *DefaultProvider) Load(ctx context.Context, region string) (aws.Config, error) {
	ensureHomeDir()
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		p.setProfile("environment")
		return loadEnvironmentCredentials(region)
	}

	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	profile := os.Getenv("AWS_PROFILE")
	explicitProfile := profile != ""
	if !explicitProfile {
		profile = "weavelens"
	}
	opts = append(opts, config.WithSharedConfigProfile(profile))

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil && !explicitProfile {
		logger := p.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Warn("failed to load profile `weavelens`, trying AWS default profile", "profile", profile, "error", err)
		cfg, err = config.LoadDefaultConfig(ctx, regionOptions(region)...)
		if err == nil {
			profile = "default"
		}
	}
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

	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		cfg.BaseEndpoint = &endpoint
	}
	p.setProfile(profile)

	return cfg, nil
}

func loadEnvironmentCredentials(region string) (aws.Config, error) {
	if region == "" {
		region = regionFromEnv()
	}
	if region == "" {
		return aws.Config{}, fmt.Errorf("AWS region is not set")
	}

	cfg := aws.Config{
		Region: region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			os.Getenv("AWS_SESSION_TOKEN"),
		)),
	}
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		cfg.BaseEndpoint = &endpoint
	}
	return cfg, nil
}

func regionOptions(region string) []func(*config.LoadOptions) error {
	if region == "" {
		return nil
	}
	return []func(*config.LoadOptions) error{config.WithRegion(region)}
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

func (p *AssumeRoleProvider) Profile() string {
	if profileProvider, ok := p.base.(interface{ Profile() string }); ok {
		return profileProvider.Profile()
	}
	return ""
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
