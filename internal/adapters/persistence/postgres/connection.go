package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"orders-service/internal/config"
	"orders-service/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewPostgresConnection(cfg *config.DatabaseConfig, logger logger.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	logger.Info("PostgreSQL connection established",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
		"max_open_conns", cfg.MaxOpenConns)

	return &PostgresDB{
		db:     db,
		logger: logger.With("component", "postgres"),
	}, nil
}

func (p *PostgresDB) DB() *sqlx.DB {
	return p.db
}

func (p *PostgresDB) Close() error {
	p.logger.Info("Closing PostgreSQL connection")
	return p.db.Close()
}

// Health check implementation
func (p *PostgresDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var result int
	err := p.db.GetContext(ctx, &result, "SELECT 1")
	if err != nil {
		p.logger.Error("PostgreSQL health check failed", "error", err)
		return fmt.Errorf("postgres health check failed: %w", err)
	}

	return nil
}

// Transaction helper
func (p *PostgresDB) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				p.logger.Error("Failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
