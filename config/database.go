package config

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetDSN(config *Config) string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DBConfig.Username, 
		config.DBConfig.Password,
		config.DBConfig.Host,
		config.DBConfig.Port,
		config.DBConfig.Name,
	)
	return dsn
}

func ConnectDatabase(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	dsn := GetDSN(cfg)
	pgxConfig, err  := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Sprintf("failed parse config: %v", err))
	}
	pgxConfig.MaxConnIdleTime = 10 * time.Minute
	pgxConfig.MaxConnLifetime = 30 * time.Minute
	pgxConfig.MaxConns = 25
	pgxConfig.MinConns = 5

	pgxPool, err  := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("Error connect database with config %w", err)
	}
	err = pgxPool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error when trying to Ping datase %w", err)
	}
	return pgxPool, nil
}

func CloseDatabase(pgx *pgxpool.Pool) {
	pgx.Close()
	slog.Info("database successfuly closed")
}