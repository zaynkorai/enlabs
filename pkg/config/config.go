package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	AppPort    string `mapstructure:"APP_PORT"`
	TimeZone   string `mapstructure:"TIME_ZONE"`
	SSLMode    string `mapstructure:"SSL_MODE"`
}

func LoadConfig() (*Config, error) {

	_ = godotenv.Load() // For local development, but env vars take precedence in production

	viper.SetDefault("APP_PORT", "8089")
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.DBHost == "" {
		cfg.DBHost = os.Getenv("DB_HOST")
		if cfg.DBHost == "" {
			return nil, fmt.Errorf("DB_HOST environment variable not set")
		}
	}
	if cfg.DBPort == "" {
		cfg.DBPort = os.Getenv("DB_PORT")
		if cfg.DBPort == "" {
			return nil, fmt.Errorf("DB_PORT environment variable not set")
		}
	}
	if cfg.DBUser == "" {
		cfg.DBUser = os.Getenv("DB_USER")
		if cfg.DBUser == "" {
			return nil, fmt.Errorf("DB_USER environment variable not set")
		}
	}
	if cfg.DBPassword == "" {
		cfg.DBPassword = os.Getenv("DB_PASSWORD")
		if cfg.DBPassword == "" {
			return nil, fmt.Errorf("DB_PASSWORD environment variable not set")
		}
	}
	if cfg.DBName == "" {
		cfg.DBName = os.Getenv("DB_NAME")
		if cfg.DBName == "" {
			return nil, fmt.Errorf("DB_NAME environment variable not set")
		}
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = os.Getenv("SSL_MODE")
		if cfg.SSLMode == "" {
			cfg.SSLMode = "disable"
		}
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = os.Getenv("TIME_ZONE")
		if cfg.TimeZone == "" {
			return nil, fmt.Errorf("TIME_ZONE environment variable not set")
		}
	}

	return &cfg, nil
}
