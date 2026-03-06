package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.redsock.ru/mead/internal/auth"
	"go.redsock.ru/mead/internal/config"
	"go.redsock.ru/mead/internal/server"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start application")
	}
}

func run() error {
	cfg, err := config.Load("./config/config.json")
	if err != nil {
		return rerrors.Wrap(err, "Load config")
	}

	// --- Credential store ---
	store := auth.NewCredentialStore()
	for _, cr := range cfg.Credentials {
		err = store.Add(cr.Username, cr.Password)
		if err != nil {
			return rerrors.Wrap(err, "Add credential", cr.Username)
		}
	}
	log.Info().Msg("Credentials loaded")

	// --- Build server ---
	srv, err := server.New(cfg, store)
	if err != nil {
		return rerrors.Wrap(err, "create server")
	}

	// --- Graceful shutdown on SIGINT/SIGTERM ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Info().
			Str("signal", sig.String()).
			Msg("received signal, shutting down")
		cancel()
		// Give active connections 30 s to finish.
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}()

	return srv.ListenAndServe(ctx)
}
