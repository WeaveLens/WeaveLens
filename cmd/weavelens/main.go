package main

import (
	"context"
	"os"

	"github.com/elip/WeaveLens/internal/application"
)

func main() {
	cfg, err := application.Load()
	if err != nil {
		panic(err)
	}

	app := application.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Run(ctx); err != nil {
		os.Exit(1)
	}
}
