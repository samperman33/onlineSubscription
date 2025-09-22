package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"database"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"SERVER_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env:"SERVER_TIMEOUT" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEUOT" env-default:"60s"`
}

type Database struct {
	Host    string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port    int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5435"`
	Name    string `yaml:"name" env:"POSTGRES_DB" env-default:"onlinesubdb"`
	User    string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
	Pass    string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"password"`
	SSLMode string `yaml:"sslmode" env:"POSTGRES_SSLMODE" env-default:"disable"`
}

func LoadCfg() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("Error opening config file: %s", err)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	return &cfg
}
