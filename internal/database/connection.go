package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/terabiome/terabiome/internal/config"
)

type ConnectionType string

const (
	ConnectionTypePostgresql ConnectionType = "postgresql"
)

type Manager struct {
	pool *pgxpool.Pool
}

var (
	globalManager *Manager
	initOnce      sync.Once
)

func Init() error {
	var err error
	initOnce.Do(func() {
		cfg := config.Get().Database
		if cfg.Type != string(ConnectionTypePostgresql) {
			err = fmt.Errorf("unsupported database type: %s", cfg.Type)
			return
		}

		poolCfg, parseErr := pgxpool.ParseConfig(cfg.DSN)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse PGX config: %w", parseErr)
			return
		}

		// Edge-optimized defaults
		poolCfg.MinConns = 1
		poolCfg.MaxConns = 5 // Slightly higher for concurrent workflows
		poolCfg.MaxConnLifetime = time.Hour
		poolCfg.MaxConnIdleTime = 30 * time.Minute

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pool, poolErr := pgxpool.NewWithConfig(ctx, poolCfg)
		if poolErr != nil {
			err = fmt.Errorf("failed to create PGX pool: %w", poolErr)
			return
		}

		if pingErr := pool.Ping(ctx); pingErr != nil {
			pool.Close()
			err = fmt.Errorf("failed to ping database: %w", pingErr)
			return
		}

		globalManager = &Manager{pool: pool}
	})
	return err
}

func GetPool() *pgxpool.Pool {
	if globalManager == nil {
		panic("database not initialized. Call database.Init() first")
	}
	return globalManager.pool
}

func Close() {
	if globalManager != nil && globalManager.pool != nil {
		globalManager.pool.Close()
	}
}
