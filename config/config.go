package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	Expiration int
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/payments?sslmode=disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
			Expiration: getEnvAsInt("JWT_EXPIRATION", 3600),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-jwt-key-change-this-in-production" {
		if c.Server.Env == "production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
