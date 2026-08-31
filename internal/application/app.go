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
	"github.com/elip/WeaveLens/internal/infrastructure/aws"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
	"github.com/elip/WeaveLens/internal/infrastructure/aws/credential"
	"github.com/elip/WeaveLens/internal/infrastructure/nats"
	"github.com/elip/WeaveLens/internal/transport"
)

type App struct {
	config   *Config
	logger   *slog.Logger
	eventBus *nats.EventBus
	scanner  *aws.AWSScanner
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

	provider := credential.Provider(&credential.DefaultProvider{})
	if cfg.AWSRoleARN != "" {
		provider = credential.NewAssumeRoleProvider(provider, credential.AssumeRoleConfig{
			RoleARN:      cfg.AWSRoleARN,
			SessionName:  cfg.AWSRoleSessionName,
			ExternalID:   cfg.AWSExternalID,
		})
	}

	var scanner *aws.AWSScanner
	if cfg.AWSRegion != "" {
		awsCfg, err := provider.Load(context.Background(), cfg.AWSRegion)
		if err != nil {
			logger.Error("failed to load AWS config", "error", err)
			os.Exit(1)
		}

		identity, err := credential.VerifyIdentity(context.Background(), awsCfg)
		if err != nil {
			logger.Error("failed to verify AWS identity", "error", err)
			os.Exit(1)
		}

		logger.Info("AWS identity verified",
			"accountID", identity.AccountID,
			"arn", identity.ARN,
			"userID", identity.UserID,
		)

		factory := client.NewFactory()
		clients := factory.BuildClients(awsCfg)
		scanner = aws.NewAWSScanner(clients)
	}

	return &App{
		config:   cfg,
		logger:   logger,
		eventBus: eventBus,
		scanner:  scanner,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting WeaveLens", "env", a.config.Env, "port", a.config.ServerPort)

	discoveryService := service.NewDiscoveryService(a.eventBus, a.logger, a.scanner)
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
