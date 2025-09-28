package ports

import (
	"context"
	"database/sql"
)

// Database transaction interface
type Transactor interface {
	WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error
}

// Message queue interfaces
type MessagePublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, message interface{}) error
}

type MessageConsumer interface {
	Consume(ctx context.Context, queue string, handler MessageHandler) error
}

type MessageHandler func(ctx context.Context, message []byte) error

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
