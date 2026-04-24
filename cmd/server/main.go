// @title           QueueForge API
// @version         1.0
// @description     Distributed job processing system
// @host            localhost:8080
// @BasePath        /
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ulasdursun/queueforge/config"
	api "github.com/ulasdursun/queueforge/internal/api"
	"github.com/ulasdursun/queueforge/internal/metrics"
	"github.com/ulasdursun/queueforge/internal/queue"
	"github.com/ulasdursun/queueforge/internal/repository"
	"github.com/ulasdursun/queueforge/internal/worker"
)

func main() {
	// Structured logging setup
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Config
	cfg := config.Load()

	// Prometheus metrics
	metrics.Register()

	// PostgreSQL
	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()
	log.Info().Msg("Connected to PostgreSQL")

	// Redis
	redisClient, err := queue.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Msg("Connected to Redis")

	// Repository + Queue
	jobRepo := repository.NewJobRepository(db)
	redisQueue := queue.NewRedisQueue(redisClient)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Worker pool — runs in background
	pool := worker.NewPool(cfg.WorkerCount, redisQueue, jobRepo)
	go pool.Start(ctx)

	// HTTP server
	router := api.NewRouter(jobRepo, redisQueue, cfg.MaxRetries)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in background
	go func() {
		log.Info().Msg(fmt.Sprintf("QueueForge API started at http://127.0.0.1:%s", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for SIGTERM or SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info().Msg("Shutdown signal received")

	// Cancel worker context
	cancel()

	// Graceful HTTP shutdown — 10s timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("QueueForge stopped cleanly")
}
