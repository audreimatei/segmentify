package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `env:"ENV" env-required:"true"`
	PostgresURL string `env:"POSTGRES_URL" env-required:"true"`
	HTTPServer
}

type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-required:"true"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-required:"true"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-required:"true"`
}

func MustLoad() *Config {
	env := os.Getenv("ENV")
	if env == "" {
		log.Fatal("ENV is not set")
	}

	var envFiles = map[string]string{
		"test": "configs/test.env",
		"dev":  "configs/dev.env",
	}

	fileName, exists := envFiles[env]
	if !exists {
		log.Fatalf("unknown ENV mode: %s", env)
	}

	err := godotenv.Load(fileName)
	if err != nil {
		log.Fatalf("failed to load .env file: %s", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read .env file: %s", err)
	}

	return &cfg
}
