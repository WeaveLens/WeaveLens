package credential

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func TestDefaultProviderLoad(t *testing.T) {
	os.Setenv("AWS_REGION", "us-west-2")
	defer os.Unsetenv("AWS_REGION")

	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")
	defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	provider := &DefaultProvider{}
	cfg, err := provider.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("DefaultProvider.Load() error = %v", err)
	}

	if cfg.Region != "us-west-2" {
		t.Errorf("DefaultProvider.Load() Region = %v, want us-west-2", cfg.Region)
	}
}

func TestDefaultProviderLoadRegionOverride(t *testing.T) {
	os.Setenv("AWS_REGION", "us-west-2")
	defer os.Unsetenv("AWS_REGION")

	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")
	defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	provider := &DefaultProvider{}
	cfg, err := provider.Load(context.Background(), "eu-central-1")
	if err != nil {
		t.Fatalf("DefaultProvider.Load() error = %v", err)
	}

	if cfg.Region != "eu-central-1" {
		t.Errorf("DefaultProvider.Load() Region = %v, want eu-central-1", cfg.Region)
	}
}

func TestDefaultProviderLoadMissingRegion(t *testing.T) {
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("AWS_ACCESS_KEY_ID")

	provider := &DefaultProvider{}
	_, err := provider.Load(context.Background(), "")
	if err == nil {
		t.Error("DefaultProvider.Load() expected error for missing region")
	}
}

func TestAssumeRoleProviderEmptyARN(t *testing.T) {
	base := &DefaultProvider{}
	provider := NewAssumeRoleProvider(base, AssumeRoleConfig{
		RoleARN: "",
	})

	cfg, err := provider.Load(context.Background(), "us-east-1")
	if err != nil {
		t.Fatalf("AssumeRoleProvider.Load() error = %v", err)
	}

	if cfg.Region != "us-east-1" {
		t.Errorf("AssumeRoleProvider.Load() Region = %v, want us-east-1", cfg.Region)
	}
}

func TestAssumeRoleProviderInvalidARN(t *testing.T) {
	base := &DefaultProvider{}
	provider := NewAssumeRoleProvider(base, AssumeRoleConfig{
		RoleARN: "not-an-arn",
	})

	_, err := provider.Load(context.Background(), "us-east-1")
	if err == nil {
		t.Error("AssumeRoleProvider.Load() expected error for invalid ARN")
	}
	if !errors.Is(err, ErrInvalidRoleARN) {
		t.Errorf("AssumeRoleProvider.Load() error = %v, want %v", err, ErrInvalidRoleARN)
	}
}

func TestAssumeRoleProviderWithMockSTS(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		action := r.FormValue("Action")

		switch action {
		case "AssumeRole":
			w.Header().Set("Content-Type", "text/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0"?>
<AssumeRoleResponse>
  <AssumeRoleResult>
    <Credentials>
      <AccessKeyId>ASIATESTKEY</AccessKeyId>
      <SecretAccessKey>testsecretkey</SecretAccessKey>
      <SessionToken>testsessiontoken</SessionToken>
      <Expiration>` + time.Now().Add(1*time.Hour).Format(time.RFC3339) + `</Expiration>
    </Credentials>
  </AssumeRoleResult>
</AssumeRoleResponse>`))
		case "GetCallerIdentity":
			w.Header().Set("Content-Type", "text/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0"?>
<GetCallerIdentityResponse>
  <GetCallerIdentityResult>
    <Account>123456789012</Account>
    <Arn>arn:aws:iam::123456789012:role/test</Arn>
    <UserId>AKIAIOSFODNN7EXAMPLE</UserId>
  </GetCallerIdentityResult>
</GetCallerIdentityResponse>`))
		default:
			t.Errorf("unexpected STS action: %s", action)
		}
	}))
	defer mockServer.Close()

	baseCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "test"))),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: mockServer.URL}, nil
		})),
	)
	if err != nil {
		t.Fatalf("failed to load base config: %v", err)
	}

	baseProvider := &staticProvider{cfg: baseCfg}
	provider := NewAssumeRoleProvider(baseProvider, AssumeRoleConfig{
		RoleARN:     "arn:aws:iam::123456789012:role/test",
		SessionName: "test-session",
	})

	cfg, err := provider.Load(context.Background(), "us-east-1")
	if err != nil {
		t.Fatalf("AssumeRoleProvider.Load() error = %v", err)
	}

	if cfg.Region != "us-east-1" {
		t.Errorf("AssumeRoleProvider.Load() Region = %v, want us-east-1", cfg.Region)
	}

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve credentials: %v", err)
	}

	if creds.AccessKeyID != "ASIATESTKEY" {
		t.Errorf("AccessKeyID = %v, want ASIATESTKEY", creds.AccessKeyID)
	}
	if creds.SecretAccessKey != "testsecretkey" {
		t.Errorf("SecretAccessKey = %v, want testsecretkey", creds.SecretAccessKey)
	}
	if creds.SessionToken != "testsessiontoken" {
		t.Errorf("SessionToken = %v, want testsessiontoken", creds.SessionToken)
	}
}

func TestVerifyIdentity(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		action := r.FormValue("Action")

		if action == "GetCallerIdentity" {
			w.Header().Set("Content-Type", "text/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0"?>
<GetCallerIdentityResponse>
  <GetCallerIdentityResult>
    <Account>123456789012</Account>
    <Arn>arn:aws:iam::123456789012:user/test-user</Arn>
    <UserId>AKIAIOSFODNN7EXAMPLE</UserId>
  </GetCallerIdentityResult>
</GetCallerIdentityResponse>`))
		}
	}))
	defer mockServer.Close()

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "test"))),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: mockServer.URL}, nil
		})),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	identity, err := VerifyIdentity(context.Background(), cfg)
	if err != nil {
		t.Fatalf("VerifyIdentity() error = %v", err)
	}

	if identity.AccountID != "123456789012" {
		t.Errorf("VerifyIdentity() AccountID = %v, want 123456789012", identity.AccountID)
	}
	if identity.ARN != "arn:aws:iam::123456789012:user/test-user" {
		t.Errorf("VerifyIdentity() ARN = %v, want arn:aws:iam::123456789012:user/test-user", identity.ARN)
	}
	if identity.UserID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("VerifyIdentity() UserID = %v, want AKIAIOSFODNN7EXAMPLE", identity.UserID)
	}
}

func TestDefaultProviderLoadWithSharedConfigProfile(t *testing.T) {
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.Mkdir(awsDir, 0700); err != nil {
		t.Fatalf("failed to create .aws dir: %v", err)
	}

	credsFile := filepath.Join(awsDir, "credentials")
	credsContent := "[test-profile]\n" +
		"aws_access_key_id = AKIAIOSFODNN7EXAMPLE\n" +
		"aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n"
	if err := os.WriteFile(credsFile, []byte(credsContent), 0600); err != nil {
		t.Fatalf("failed to write credentials file: %v", err)
	}

	configFile := filepath.Join(awsDir, "config")
	configContent := "[profile test-profile]\n" +
		"region = us-east-1\n"
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	envVarsToRestore := []string{
		"HOME", "USERPROFILE",
		"AWS_PROFILE", "AWS_DEFAULT_PROFILE",
		"AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY",
		"AWS_SECRET_ACCESS_KEY", "AWS_SECRET_KEY",
		"AWS_SESSION_TOKEN",
		"AWS_REGION", "AWS_DEFAULT_REGION",
		"AWS_CONFIG_FILE", "AWS_SHARED_CREDENTIALS_FILE",
	}
	savedValues := make(map[string]string, len(envVarsToRestore))
	savedExists := make(map[string]bool, len(envVarsToRestore))
	for _, key := range envVarsToRestore {
		savedValues[key], savedExists[key] = os.LookupEnv(key)
	}
	t.Cleanup(func() {
		for _, key := range envVarsToRestore {
			if savedExists[key] {
				os.Setenv(key, savedValues[key])
			} else {
				os.Unsetenv(key)
			}
		}
	})

	os.Setenv("HOME", tmpDir)
	os.Unsetenv("USERPROFILE")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_ACCESS_KEY")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SECRET_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")
	os.Setenv("AWS_PROFILE", "test-profile")
	os.Setenv("AWS_CONFIG_FILE", configFile)
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", credsFile)

	provider := &DefaultProvider{}
	cfg, err := provider.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("DefaultProvider.Load() error = %v", err)
	}

	if cfg.Region != "us-east-1" {
		t.Errorf("DefaultProvider.Load() Region = %v, want us-east-1", cfg.Region)
	}

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve credentials: %v", err)
	}

	if creds.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("AccessKeyID = %v, want AKIAIOSFODNN7EXAMPLE", creds.AccessKeyID)
	}
	if creds.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("SecretAccessKey = %v, want wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", creds.SecretAccessKey)
	}
}

func TestEnsureHomeDir(t *testing.T) {
	originalHome, homeExisted := os.LookupEnv("HOME")
	originalUserProfile, userProfileExisted := os.LookupEnv("USERPROFILE")
	originalConfigFile, configExisted := os.LookupEnv("AWS_CONFIG_FILE")
	originalCredsFile, credsExisted := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE")

	t.Cleanup(func() {
		if homeExisted {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if userProfileExisted {
			os.Setenv("USERPROFILE", originalUserProfile)
		} else {
			os.Unsetenv("USERPROFILE")
		}
		if configExisted {
			os.Setenv("AWS_CONFIG_FILE", originalConfigFile)
		} else {
			os.Unsetenv("AWS_CONFIG_FILE")
		}
		if credsExisted {
			os.Setenv("AWS_SHARED_CREDENTIALS_FILE", originalCredsFile)
		} else {
			os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
		}
	})

	home := os.Getenv("HOME")

	os.Unsetenv("HOME")
	os.Unsetenv("USERPROFILE")
	os.Unsetenv("AWS_CONFIG_FILE")
	os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")

	ensureHomeDir()

	if home != "" {
		if got := os.Getenv("HOME"); got != home {
			t.Errorf("HOME = %q, want %q", got, home)
		}
	} else {
		t.Errorf("HOME should have been resolved to a non-empty value")
	}

	if got := os.Getenv("USERPROFILE"); got == "" {
		t.Errorf("USERPROFILE should have been set")
	}

	if got := os.Getenv("AWS_CONFIG_FILE"); got == "" {
		t.Errorf("AWS_CONFIG_FILE should have been set")
	}
	if got := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); got == "" {
		t.Errorf("AWS_SHARED_CREDENTIALS_FILE should have been set")
	}
}

func TestEnsureHomeDirPreservesExistingHome(t *testing.T) {
	originalHome, homeExisted := os.LookupEnv("HOME")
	originalUserProfile, userProfileExisted := os.LookupEnv("USERPROFILE")
	originalConfigFile, configExisted := os.LookupEnv("AWS_CONFIG_FILE")
	originalCredsFile, credsExisted := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE")

	t.Cleanup(func() {
		if homeExisted {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if userProfileExisted {
			os.Setenv("USERPROFILE", originalUserProfile)
		} else {
			os.Unsetenv("USERPROFILE")
		}
		if configExisted {
			os.Setenv("AWS_CONFIG_FILE", originalConfigFile)
		} else {
			os.Unsetenv("AWS_CONFIG_FILE")
		}
		if credsExisted {
			os.Setenv("AWS_SHARED_CREDENTIALS_FILE", originalCredsFile)
		} else {
			os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
		}
	})

	tmpDir := t.TempDir()
	originalHome = tmpDir
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("USERPROFILE")
	os.Unsetenv("AWS_CONFIG_FILE")
	os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")

	ensureHomeDir()

	if got := os.Getenv("HOME"); got != tmpDir {
		t.Errorf("HOME should be preserved, got %q, want %q", got, tmpDir)
	}

	if got := os.Getenv("AWS_CONFIG_FILE"); got != "" {
		t.Errorf("AWS_CONFIG_FILE should not be set when HOME was already set, got %q", got)
	}
	if got := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); got != "" {
		t.Errorf("AWS_SHARED_CREDENTIALS_FILE should not be set when HOME was already set, got %q", got)
	}
}

func TestErrorRedaction(t *testing.T) {
	base := &DefaultProvider{}
	provider := NewAssumeRoleProvider(base, AssumeRoleConfig{
		RoleARN: "not-an-arn",
	})

	_, err := provider.Load(context.Background(), "us-east-1")
	if err == nil {
		t.Fatal("expected error")
	}

	errStr := err.Error()
	if strings.Contains(errStr, "secret") || strings.Contains(errStr, "password") || strings.Contains(errStr, "token") {
		t.Errorf("error contains sensitive material: %v", errStr)
	}
}

type staticProvider struct {
	cfg aws.Config
}

func (p *staticProvider) Load(ctx context.Context, region string) (aws.Config, error) {
	return p.cfg, nil
}
