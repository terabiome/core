package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/terabiome/core/internal/config"
	"github.com/terabiome/core/pkg/database/postgresql"
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

		pool, poolErr := postgresql.Open(context.Background(), postgresql.Config{DSN: cfg.DSN})
		if poolErr != nil {
			err = fmt.Errorf("failed to open database: %w", poolErr)
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
