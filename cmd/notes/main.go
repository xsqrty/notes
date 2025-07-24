package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xsqrty/notes/internal/api/prometheus"
	"github.com/xsqrty/notes/internal/api/rest"
	"github.com/xsqrty/notes/internal/api/swag"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/internal/logger"
	"github.com/xsqrty/notes/pkg/httputil/httpgs"
	"github.com/xsqrty/op/db/postgres"
)

func main() {
	start := time.Now()
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("config loader: %w", err))
	}

	if cfg.PrintVersion() {
		fmt.Printf("%s version %s\n", cfg.AppName, cfg.Version)
		os.Exit(0)
	}

	if cfg.PrintHelp() {
		fmt.Printf("%s\n", cfg.Help())
		os.Exit(0)
	}

	log, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		panic(fmt.Errorf("logger loader: %w", err))
	}

	pool, err := postgres.Open(ctx, cfg.DB.DSN)
	if err != nil {
		panic(fmt.Errorf("postgres connection: %w", err))
	}
	defer func() {
		if err := pool.Close(); err != nil {
			panic(fmt.Errorf("postgres close connection error: %w", err))
		}
	}()

	deps := app.NewDeps(cfg, log, pool)
	defer func() {
		if err := deps.Close(); err != nil {
			panic(fmt.Errorf("close deps error: %w", err))
		}
	}()

	rest := rest.NewRest(deps)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           rest.Routes(),
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = httpgs.NewGracefulShutdown(ctx).
		OnMessage(func(name, message string) {
			log.Info().Msg(fmt.Sprintf("%s: %s", name, message))
		}).
		OnError(func(name string, err error) {
			log.Error().Msg(fmt.Sprintf("%s: %s", name, err.Error()))
		}).
		Register(
			"Rest",
			server,
			httpgs.WithShutdownTimeout(cfg.Server.ShutdownTimeout),
			httpgs.WithHighPriority(),
		).
		Register("Swag", swag.NewSwagServer(cfg.Swag), httpgs.WithShutdownTimeout(cfg.Swag.ShutdownTimeout)).
		Register("Prom", prometheus.NewPrometheusServer(cfg.Metrics), httpgs.WithShutdownTimeout(cfg.Metrics.ShutdownTimeout)).
		ListenAndServe()
	if err != nil {
		log.Error().Err(err).Msg("Graceful shutdown error")
	}

	log.Info().Str("worked_at", time.Since(start).String()).Msg("Application stopped")
}
