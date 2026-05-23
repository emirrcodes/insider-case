package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerPort           string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	SimulationSeed       int64
	PredictionSeed       int64
	PredictionIterations int
}

func Load() Config {
	return Config{
		ServerPort:           env("SERVER_PORT", "8080"),
		DBHost:               env("DB_HOST", "localhost"),
		DBPort:               env("DB_PORT", "5432"),
		DBUser:               env("DB_USER", "insider"),
		DBPassword:           env("DB_PASSWORD", "insider"),
		DBName:               env("DB_NAME", "insider_league"),
		DBSSLMode:            env("DB_SSLMODE", "disable"),
		SimulationSeed:       envInt64("SIMULATION_SEED", 20260518),
		PredictionSeed:       envInt64("PREDICTION_SEED", 42024),
		PredictionIterations: envInt("PREDICTION_ITERATIONS", 10000),
	}
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
