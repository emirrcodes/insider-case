package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ahmetemirarslan/insider-backend-case/internal/config"
	httpapi "github.com/ahmetemirarslan/insider-backend-case/internal/http"
	"github.com/ahmetemirarslan/insider-backend-case/internal/repository/postgres"
	"github.com/ahmetemirarslan/insider-backend-case/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	pool, err := connectWithRetry(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewRepository(pool)
	leagueService := service.NewLeagueService(
		repo,
		service.DeterministicSimulator{Seed: cfg.SimulationSeed},
		service.MonteCarloPredictor{
			Iterations: cfg.PredictionIterations,
			Seed:       cfg.PredictionSeed,
		},
	)

	router := gin.Default()
	httpapi.NewHandler(leagueService).RegisterRoutes(router)

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server listening on :%s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func connectWithRetry(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	var lastErr error
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				lastErr = pingErr
				pool.Close()
			}
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return nil, lastErr
		case <-ticker.C:
		}
	}
}
