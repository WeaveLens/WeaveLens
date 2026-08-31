package credential

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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
