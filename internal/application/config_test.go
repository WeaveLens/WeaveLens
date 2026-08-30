package application_test

import (
	"os"
	"testing"

	"github.com/elip/WeaveLens/internal/application"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("ENV")
	os.Unsetenv("LOG_LEVEL")

	cfg, err := application.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ServerPort != "8080" {
		t.Errorf("ServerPort = %v, want 8080", cfg.ServerPort)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %v, want development", cfg.Env)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("ENV", "production")
	os.Setenv("LOG_LEVEL", "debug")

	cfg, err := application.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ServerPort != "9090" {
		t.Errorf("ServerPort = %v, want 9090", cfg.ServerPort)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %v, want production", cfg.Env)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}

	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("ENV")
	os.Unsetenv("LOG_LEVEL")
}
