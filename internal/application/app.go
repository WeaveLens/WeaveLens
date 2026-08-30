package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elip/WeaveLens/internal/transport"
)

type App struct {
	config *Config
	logger *slog.Logger
}

func New(cfg *Config) *App {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &App{
		config: cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting WeaveLens", "env", a.config.Env, "port", a.config.ServerPort)

	mux := transport.NewRouter()
	server, err := transport.StartServer(":"+a.config.ServerPort, mux)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		a.logger.Info("context cancelled, shutting down")
	case sig := <-sigCh:
		a.logger.Info("received signal", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("server shutdown error", "error", err)
		return err
	}

	a.logger.Info("server stopped gracefully")
	return nil
}
