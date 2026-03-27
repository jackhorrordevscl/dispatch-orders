package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool crea un pool de conexiones a PostgreSQL
func NewPool(ctx context.Context, config *Config) (*pgxpool.Pool, error) {
    poolConfig, err := pgxpool.ParseConfig(config.DSN())
    if err != nil {
        return nil, fmt.Errorf("failed to parse pool config: %w", err)
    }
    
    // Configuración del pool
    poolConfig.MaxConns = 25
    poolConfig.MinConns = 5
    poolConfig.MaxConnLifetime = time.Hour
    poolConfig.MaxConnIdleTime = 30 * time.Minute
    poolConfig.HealthCheckPeriod = time.Minute
    
    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create pool: %w", err)
    }
    
    // Verificar conexión
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return pool, nil
}