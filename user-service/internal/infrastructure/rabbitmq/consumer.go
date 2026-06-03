package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DoMinhHHung/user-service/internal/config"
	"github.com/DoMinhHHung/user-service/internal/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type UserCreatedEvent struct {
	Event  string `json:"event"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type EventHandler interface {
	HandleUserCreated(ctx context.Context, event UserCreatedEvent) error
}

type Consumer struct {
	conn    *amqp.Connection
	cfg     config.RabbitMQConfig
	handler EventHandler
	log     *logger.Logger
}

func NewConsumer(cfg config.RabbitMQConfig, handler EventHandler, log *logger.Logger) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	return &Consumer{conn: conn, cfg: cfg, handler: handler, log: log}, nil
}

func (c *Consumer) Close() {
	if c.conn != nil && !c.conn.IsClosed() {
		c.conn.Close()
	}
}

func (c *Consumer) Setup() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		c.cfg.Exchange, "topic", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	dlxName := c.cfg.Exchange + ".dlx"
	if err := ch.ExchangeDeclare(
		dlxName, "direct", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare dlx: %w", err)
	}

	if _, err := ch.QueueDeclare(
		c.cfg.DLQUserCreated, true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	if err := ch.QueueBind(
		c.cfg.DLQUserCreated, c.cfg.DLQUserCreated, dlxName, false, nil,
	); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	mainArgs := amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": c.cfg.DLQUserCreated,
		"x-message-ttl":             int32(30000),
	}
	if _, err := ch.QueueDeclare(
		c.cfg.QueueUserCreated, true, false, false, false, mainArgs,
	); err != nil {
		return fmt.Errorf("declare main queue: %w", err)
	}
	if err := ch.QueueBind(
		c.cfg.QueueUserCreated, "user.created", c.cfg.Exchange, false, nil,
	); err != nil {
		return fmt.Errorf("bind main queue: %w", err)
	}

	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	msgs, err := ch.Consume(
		c.cfg.QueueUserCreated,
		"user-service-consumer",
		false,
		false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	c.log.Info("rabbitmq consumer started", "queue", c.cfg.QueueUserCreated)

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.log.Info("rabbitmq consumer stopped")
				ch.Close()
				return

			case msg, ok := <-msgs:
				if !ok {
					c.log.Warn("rabbitmq channel closed")
					return
				}
				c.processMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	log := c.log.With(
		"message_id", msg.MessageId,
		"routing_key", msg.RoutingKey,
	)

	var event UserCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error("invalid message format, sending to DLQ", "error", err)
		_ = msg.Reject(false)
		return
	}

	if event.UserID == "" || event.Email == "" {
		log.Error("invalid event payload", "event", event)
		_ = msg.Reject(false)
		return
	}

	retryCount := getRetryCount(msg.Headers)
	log = log.With("retry_count", retryCount, "user_id", event.UserID)

	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := c.handler.HandleUserCreated(processCtx, event); err != nil {
		log.Error("failed to handle user.created", "error", err)

		if retryCount >= c.cfg.MaxRetry {
			log.Error("max retry exceeded, sending to DLQ", "user_id", event.UserID)
			_ = msg.Reject(false)
			return
		}

		delay := time.Duration(1<<uint(retryCount)) * time.Second
		time.Sleep(delay)
		_ = msg.Nack(false, true)
		return
	}

	_ = msg.Ack(false)
	log.Info("user.created event processed successfully")
}

func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	deaths, ok := headers["x-death"].([]interface{})
	if !ok || len(deaths) == 0 {
		return 0
	}
	if death, ok := deaths[0].(amqp.Table); ok {
		if count, ok := death["count"].(int64); ok {
			return int(count)
		}
	}
	return 0
}
