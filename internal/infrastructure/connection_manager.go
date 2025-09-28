package infrastructure

import (
	"context"
	"fmt"

	"orders-service/internal/adapters/messaging/rabbitmq"
	"orders-service/internal/adapters/persistence/postgres"
	"orders-service/internal/application/ports"
	"orders-service/internal/config"
	"orders-service/pkg/logger"
)

type DatabaseConnections struct {
	Postgres *postgres.PostgresDB
	RabbitMQ *rabbitmq.RabbitMQClient
	logger   logger.Logger
}

func NewDatabaseConnections(cfg *config.Config, logger logger.Logger) (*DatabaseConnections, error) {
	log := logger.With("component", "database_connections")

	// PostgreSQL connection
	log.Info("Connecting to PostgreSQL...")
	pg, err := postgres.NewPostgresConnection(&cfg.Database, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// RabbitMQ connection
	log.Info("Connecting to RabbitMQ...")
	rmq, err := rabbitmq.NewRabbitMQConnection(&cfg.RabbitMQ, logger)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	log.Info("All database connections established successfully")

	return &DatabaseConnections{
		Postgres: pg,
		RabbitMQ: rmq,
		logger:   log,
	}, nil
}

func (d *DatabaseConnections) Close() error {
	d.logger.Info("Closing all database connections...")

	var errs []error

	if err := d.RabbitMQ.Close(); err != nil {
		errs = append(errs, fmt.Errorf("rabbitmq close error: %w", err))
	}

	if err := d.Postgres.Close(); err != nil {
		errs = append(errs, fmt.Errorf("postgres close error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	d.logger.Info("All database connections closed successfully")
	return nil
}

func (d *DatabaseConnections) HealthCheck(ctx context.Context) map[string]error {
	checks := make(map[string]error)

	checks["postgres"] = d.Postgres.HealthCheck(ctx)
	checks["rabbitmq"] = d.RabbitMQ.HealthCheck(ctx)

	return checks
}

func (d *DatabaseConnections) GetMessagePublisher() ports.MessagePublisher {
	return d.RabbitMQ
}

func (d *DatabaseConnections) GetMessageConsumer() ports.MessageConsumer {
	return d.RabbitMQ
}
