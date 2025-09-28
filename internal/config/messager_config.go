package config

import (
	"time"

	"github.com/spf13/viper"
)

type RabbitMQConfig struct {
	URL              string        `mapstructure:"url"`
	Host             string        `mapstructure:"host"`
	Port             string        `mapstructure:"port"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	VHost            string        `mapstructure:"vhost"`
	ConnectionName   string        `mapstructure:"connection_name"`
	HeartbeatTimeout time.Duration `mapstructure:"heartbeat_timeout"`
	MaxRetries       int           `mapstructure:"max_retries"`
}

func RabbitMQDefaults(v *viper.Viper) {
	v.SetDefault("rabbitmq.host", "localhost")
	v.SetDefault("rabbitmq.port", "5672")
	v.SetDefault("rabbitmq.username", "guest")
	v.SetDefault("rabbitmq.password", "guest")
	v.SetDefault("rabbitmq.vhost", "/")
	v.SetDefault("rabbitmq.connection_name", "orders-service")
	v.SetDefault("rabbitmq.heartbeat_timeout", 60*time.Second)
	v.SetDefault("rabbitmq.max_retries", 3)
}
