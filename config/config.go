package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Kafka     KafkaConfig
	Treasury  TreasuryConfig
	Server    ServerConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
	Metrics   MetricsConfig
}

type AppConfig struct {
	Name     string
	Env      string
	Port     int
	Debug    bool
	LogLevel string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type KafkaConfig struct {
	Brokers          []string
	GroupID          string
	TransactionTopic string
	ClientID         string
	AutoOffsetReset  string
	EnableAutoCommit bool
}

type TreasuryConfig struct {
	APIURL        string
	Timeout       time.Duration
	CacheDuration time.Duration
}

type ServerConfig struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type SecurityConfig struct {
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
	CORSMaxAge         int
}

type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

type MetricsConfig struct {
	Enabled bool
	Port    int
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &config, nil
}

// ValidateConfig performs validation of the loaded configuration
func (c *Config) Validate() error {
	if c.App.Port <= 0 {
		return fmt.Errorf("invalid port number: %d", c.App.Port)
	}

	if c.Database.MaxOpenConns < c.Database.MaxIdleConns {
		return fmt.Errorf("maxOpenConns must be greater than or equal to maxIdleConns")
	}

	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker must be configured")
	}

	return nil
}
