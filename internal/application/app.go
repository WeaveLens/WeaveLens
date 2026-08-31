package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elip/WeaveLens/internal/application/service"
	"github.com/elip/WeaveLens/internal/infrastructure/nats"
	"github.com/elip/WeaveLens/internal/transport"
)

type App struct {
	config   *Config
	logger   *slog.Logger
	eventBus *nats.EventBus
}

func New(cfg *Config) *App {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	natsClient, err := nats.Connect(context.Background(), &nats.Config{
		URL:           cfg.NATSURL,
		StreamName:    "weavelens",
		MaxAge:        24 * time.Hour,
		MaxMsgs:       1000000,
		DurablePrefix: "weavelens",
		AckWait:       30 * time.Second,
		MaxDeliver:    1,
		Backoff:       []time.Duration{time.Second, 5 * time.Second, 15 * time.Second},
	})
	if err != nil {
		logger.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}

	publisher := nats.NewJetStreamPublisher(natsClient)
	subscriber := nats.NewJetStreamSubscriber(natsClient)
	eventBus := nats.NewEventBus(publisher, subscriber)

	return &App{
		config:   cfg,
		logger:   logger,
		eventBus: eventBus,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting WeaveLens", "env", a.config.Env, "port", a.config.ServerPort)

	discoveryService := service.NewDiscoveryService(a.eventBus, a.logger)
	graphService := service.NewGraphService(a.eventBus, a.logger)

	mux := transport.NewRouter(discoveryService, graphService)
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

	if err := a.eventBus.Close(); err != nil {
		a.logger.Error("event bus shutdown error", "error", err)
	}

	a.logger.Info("server stopped gracefully")
	return nil
}
