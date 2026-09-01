package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/elip/WeaveLens/internal/application/service"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/credential"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/discovery"
	"github.com/elip/WeaveLens/internal/infrastructure/nats"
	"github.com/elip/WeaveLens/internal/transport"
)

type App struct {
	config           *Config
	logger           *slog.Logger
	eventBus         *nats.EventBus
	discovery        discovery.ResourceDiscovery
	identity         *credential.Identity
	region           string
	credentialSource string
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func New(cfg *Config) *App {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
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

	provider := credential.Provider(&credential.DefaultProvider{})
	credentialSource := "default"
	if cfg.AWSRoleARN != "" {
		provider = credential.NewAssumeRoleProvider(provider, credential.AssumeRoleConfig{
			RoleARN:     cfg.AWSRoleARN,
			SessionName: cfg.AWSRoleSessionName,
			ExternalID:  cfg.AWSExternalID,
		})
		credentialSource = "assume_role"
	}

	var disc discovery.ResourceDiscovery
	var identity *credential.Identity
	if cfg.AWSRegion != "" {
		awsCfg, err := provider.Load(context.Background(), cfg.AWSRegion)
		if err != nil {
			logger.Error("failed to load AWS config", "error", err)
			logger.Warn("continuing without AWS connection - set AWS_REGION and credentials to enable")
		} else {
			identity, err = credential.VerifyIdentity(context.Background(), awsCfg)
			if err != nil {
				logger.Error("failed to verify AWS identity", "error", err)
				logger.Warn("continuing without AWS connection - check credentials")
			} else {
				logger.Info("AWS identity verified",
					"accountID", identity.AccountID,
					"arn", identity.ARN,
					"userID", identity.UserID,
					"credential_source", credentialSource,
				)
			}

			factory := client.NewFactory()
			clients := factory.BuildClients(awsCfg)
			disc = discovery.NewServiceFromConfig(discovery.ServiceConfigInput{
				Clients:   clients,
				Region:    cfg.AWSRegion,
				AWSConfig: awsCfg,
			})
		}
	}

	return &App{
		config:           cfg,
		logger:           logger,
		eventBus:         eventBus,
		discovery:        disc,
		identity:         identity,
		region:           cfg.AWSRegion,
		credentialSource: credentialSource,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting WeaveLens", "env", a.config.Env, "port", a.config.ServerPort)

	discoveryService := service.NewDiscoveryService(a.eventBus, a.logger, a.discovery)
	graphService := service.NewGraphServiceWithCallback(a.eventBus, a.logger, a.discovery, func(scanID string, nodeCount, edgeCount int) {
		if err := discoveryService.CompleteScan(context.Background(), scanID, nodeCount, edgeCount); err != nil {
			a.logger.Error("failed to complete scan", "error", err, "scanID", scanID)
		}
	})

	discoveryService.SetGraphService(graphService)
	exportService := service.NewExportService(graphService)

	mux := transport.NewRouter(discoveryService, graphService, a, exportService, a.logger)
	server, err := transport.StartServer(":"+a.config.ServerPort, mux, a.config.APIKey, a.logger)
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

func (a *App) GetConnectionStatus() transport.ConnectionStatus {
	if a.identity == nil {
		return transport.ConnectionStatus{
			State:   "not_connected",
			Message: "AWS credentials not configured. Set AWS_REGION and ensure credentials are available.",
		}
	}

	return transport.ConnectionStatus{
		State:            "connected",
		AccountID:        a.identity.AccountID,
		ARN:              a.identity.ARN,
		Region:           a.region,
		CredentialSource: a.credentialSource,
		Message:          "",
	}
}

// ConnectionStatus is an alias for transport.ConnectionStatus for backward compatibility.
type ConnectionStatus = transport.ConnectionStatus
