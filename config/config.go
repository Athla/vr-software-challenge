package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{App: {Env: %s, Port: %d, Debug: %v, LogLevel: %s}, "+
			"Database: {Host: %s, Port: %d, User: %s, Name: %s, SSLMode: %s}, "+
			"Kafka: {Brokers: %v, GroupID: %s, Topic: %s, ClientID: %s}}",
		c.App.Env, c.App.Port, c.App.Debug, c.App.LogLevel,
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Name, c.Database.SSLMode,
		c.Kafka.Brokers, c.Kafka.GroupID, c.Kafka.Topic, c.Kafka.ClientID,
	)
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

type AppConfig struct {
	Env      string
	Port     int
	Debug    bool
	LogLevel string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int    `default:"25"`
	MaxIdleConns    int    `default:"5"`
	ConnMaxLifetime string `default:"5m"`
}

type KafkaConfig struct {
	Brokers  []string
	GroupID  string
	Topic    string
	ClientID string
}

func Load() (*Config, error) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	config := &Config{
		App: AppConfig{
			Env:      os.Getenv("APP_ENV"),
			Port:     port,
			Debug:    debug,
			LogLevel: os.Getenv("LOG_LEVEL"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     dbPort,
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSL_MODE"),
		},
		Kafka: KafkaConfig{
			Brokers:  strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
			GroupID:  os.Getenv("KAFKA_GROUP_ID"),
			Topic:    os.Getenv("KAFKA_TOPIC"),
			ClientID: os.Getenv("KAFKA_CLIENT_ID"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.App.Port <= 0 {
		return fmt.Errorf("invalid port number: %d", c.App.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Port <= 0 {
		return fmt.Errorf("invalid database port")
	}

	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}

	if c.Kafka.GroupID == "" {
		return fmt.Errorf("Kafka group ID is required")
	}

	if c.Kafka.Topic == "" {
		return fmt.Errorf("Kafka topic is required")
	}

	return nil
}

func (cfg *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	duration, err := time.ParseDuration(cfg.ConnMaxLifetime)
	if err != nil {
		return 5 * time.Minute // default
	}
	return duration
}

func (cfg *DatabaseConfig) ConnString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)
}
