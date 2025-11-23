package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Kafka    KafkaConfig
	Server   ServerConfig
	Logging  LoggingConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type ServerConfig struct {
	Port string
}

type LoggingConfig struct {
	Level string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}, ","),
			Topic:   getEnv("KAFKA_TOPIC", "orders"),
			GroupID: getEnv("KAFKA_GROUP_ID", "order-service"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultSlice []string, sep string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultSlice
	}
	return strings.Split(value, sep)
}

func (c *DatabaseConfig) GetDBConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

func (c *Config) Validate() error {
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER environment variable is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME environment variable is required")
	}
	return nil
}
