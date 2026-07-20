package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig AppConfig
	DBConfig DBConfig
	JWTConfig *JWTConfig
}
type AppConfig struct {
	Env string
	Port string
}

type DBConfig struct {
	Host string
	Name string
	Port string
	Username string
	Password string
}

type JWTConfig struct {
	Secret    string
	AccessTTL time.Duration
	RefreshTTL time.Duration
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Notice: .env file not found, reading from system environment variables")
	}
	accessTLL, err := time.ParseDuration("15m")
	if err != nil {
		panic(fmt.Sprintf("invalid JWT accessTTL: %v", err)) 
	}
	refreshTTL, err := time.ParseDuration("168h")
	if err != nil {
		panic(fmt.Sprintf("invalid JWT refreshTTL: %v", err)) 
	}
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := requireEnv("DB_USER")
	dbPort := getEnv("DB_PORT", "5433")
	dbPass := os.Getenv("DB_PASS")
	dbName := requireEnv("DB_NAME")

	return &Config{
		AppConfig: AppConfig{
			Env: getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", ":50052"),
		},
		DBConfig: DBConfig{
			Host: dbHost,
			Name: dbName,
			Port: dbPort,
			Username: dbUser,
			Password: dbPass,
		},
		JWTConfig: &JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
			AccessTTL: accessTLL,
			RefreshTTL: refreshTTL,
		},
	}
}

func getEnv(key, fallback string) string {
	if val:= os.Getenv(key); val != "" {return  val}
	return fallback
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("system use env, require variable : %s", key))
	}
	return val
}