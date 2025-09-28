package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"orders-service/internal/application/ports"
	"orders-service/internal/config"
	"orders-service/pkg/logger"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *config.RabbitMQConfig
	logger  logger.Logger
}

func NewRabbitMQConnection(cfg *config.RabbitMQConfig, logger logger.Logger) (*RabbitMQClient, error) {
	var url string
	if cfg.URL != "" {
		url = cfg.URL
	} else {
		url = fmt.Sprintf("amqp://%s:%s@%s:%s%s",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)
	}

	conn, err := amqp091.DialConfig(url, amqp091.Config{
		Heartbeat: cfg.HeartbeatTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	logger.Info("RabbitMQ connection established",
		"host", cfg.Host,
		"port", cfg.Port,
		"vhost", cfg.VHost)

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
		config:  cfg,
		logger:  logger.With("component", "rabbitmq"),
	}, nil
}

func (r *RabbitMQClient) Close() error {
	r.logger.Info("Closing RabbitMQ connection")
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Health check implementation
func (r *RabbitMQClient) HealthCheck(ctx context.Context) error {
	if r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	if r.channel == nil {
		return fmt.Errorf("rabbitmq channel is nil")
	}
	return nil
}

// Publisher implementation
func (r *RabbitMQClient) Publish(ctx context.Context, exchange, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp091.Persistent, // Make messages persistent
		},
	)

	if err != nil {
		r.logger.Error("Failed to publish message",
			"exchange", exchange,
			"routing_key", routingKey,
			"error", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.logger.Debug("Message published",
		"exchange", exchange,
		"routing_key", routingKey)

	return nil
}

// Consumer implementation
func (r *RabbitMQClient) Consume(ctx context.Context, queue string, handler ports.MessageHandler) error {
	msgs, err := r.channel.Consume(
		queue,
		"",    // consumer
		false, // auto-ack (manual ack for reliability)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	r.logger.Info("Starting to consume messages", "queue", queue)

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				r.logger.Warn("Message channel closed")
				return fmt.Errorf("message channel closed")
			}

			err := handler(ctx, msg.Body)
			if err != nil {
				r.logger.Error("Failed to handle message",
					"queue", queue,
					"error", err)
				msg.Nack(false, true) // Requeue message
			} else {
				msg.Ack(false)
				r.logger.Debug("Message processed successfully", "queue", queue)
			}

		case <-ctx.Done():
			r.logger.Info("Context cancelled, stopping consumer", "queue", queue)
			return ctx.Err()
		}
	}
}

// Ensure RabbitMQClient implements the interfaces
var _ ports.MessagePublisher = (*RabbitMQClient)(nil)
var _ ports.MessageConsumer = (*RabbitMQClient)(nil)
