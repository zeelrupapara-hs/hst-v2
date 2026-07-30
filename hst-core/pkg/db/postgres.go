package db

import (
	"context"
	"fmt"
	"time"

	"hstcore/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB instant to pass to handlers
type PostgresDB struct {
	DB *pgxpool.Pool
}

// NewPostgresDB will return a valid connection to Postgres DB Session
func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Postgres.Dsn())
	if err != nil {
		return nil, fmt.Errorf("unable to parse postgres dsn: %w", err)
	}

	poolCfg.MaxConns = cfg.Postgres.PostgresMaxConn
	poolCfg.MinConns = cfg.Postgres.PostgresMinConn
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping postgres: %w", err)
	}

	return &PostgresDB{DB: pool}, nil
}

// Migrate will create the schema, called once at startup
func (db *PostgresDB) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := db.DB.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS hst`)
	return err
}
