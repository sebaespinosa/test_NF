package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Jaeger   JaegerConfig
	Loki     LokiConfig
	Service  ServiceConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port uint16
	Env  string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string
	Port            uint16
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	DSN             string
}

// JaegerConfig holds Jaeger tracing configuration
type JaegerConfig struct {
	AgentHost    string
	AgentPort    uint16
	SamplerType  string
	SamplerParam float64
}

// LokiConfig holds Loki logging configuration
type LokiConfig struct {
	URL string
}

// ServiceConfig holds service-related configuration
type ServiceConfig struct {
	Name    string
	Version string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: parseUint16(os.Getenv("SERVER_PORT"), 8080),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            parseUint16(os.Getenv("DB_PORT"), 5432),
			User:            getEnv("DB_USER", "irrigationuser"),
			Password:        getEnv("DB_PASSWORD", "irrigationpass"),
			Name:            getEnv("DB_NAME", "irrigation_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    parseInt(os.Getenv("DB_MAX_OPEN_CONNS"), 25),
			MaxIdleConns:    parseInt(os.Getenv("DB_MAX_IDLE_CONNS"), 5),
			ConnMaxLifetime: parseDuration(os.Getenv("DB_CONN_MAX_LIFETIME"), "5m"),
		},
		Jaeger: JaegerConfig{
			AgentHost:    getEnv("JAEGER_AGENT_HOST", "localhost"),
			AgentPort:    parseUint16(os.Getenv("JAEGER_AGENT_PORT"), 6831),
			SamplerType:  getEnv("JAEGER_SAMPLER_TYPE", "const"),
			SamplerParam: parseFloat64(os.Getenv("JAEGER_SAMPLER_PARAM"), 1.0),
		},
		Loki: LokiConfig{
			URL: getEnv("LOKI_URL", "http://localhost:3100"),
		},
		Service: ServiceConfig{
			Name:    getEnv("SERVICE_NAME", "irrigation-api"),
			Version: getEnv("SERVICE_VERSION", "0.0.1"),
		},
	}

	// Build PostgreSQL DSN
	cfg.Database.DSN = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func parseUint16(value string, defaultVal uint16) uint16 {
	if value == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return defaultVal
	}
	return uint16(parsed)
}

func parseInt(value string, defaultVal int) int {
	if value == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func parseFloat64(value string, defaultVal float64) float64 {
	if value == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func parseDuration(value string, defaultVal string) time.Duration {
	if value == "" {
		value = defaultVal
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		defaultDuration, _ := time.ParseDuration(defaultVal)
		return defaultDuration
	}
	return duration
}
