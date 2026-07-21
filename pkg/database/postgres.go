package database

import (
	"context"
	"employee_management/config"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("database: parse postgres config: %w", err)
	}
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 1
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: create postgres pool:: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
		lastErr = pool.Ping(pingCtx)
		pingCancel()
		if lastErr == nil {
			return pool, nil
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	pool.Close()
	return nil, fmt.Errorf("database: ping postgres after retires: %w", lastErr)
}
