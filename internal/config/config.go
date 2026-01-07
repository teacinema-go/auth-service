package config

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/teacinema-go/core/constants"
)

type Config struct {
	App      AppConfig      `mapstructure:",squash"`
	Postgres PostgresConfig `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
}

type AppConfig struct {
	Env  constants.Env `mapstructure:"APP_ENV" validate:"required"`
	Port int           `mapstructure:"APP_PORT" validate:"required"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"POSTGRES_HOST" validate:"required"`
	Port     int    `mapstructure:"POSTGRES_PORT" validate:"required"`
	User     string `mapstructure:"POSTGRES_USER" validate:"required"`
	Password string `mapstructure:"POSTGRES_PASSWORD" validate:"required"`
	Name     string `mapstructure:"POSTGRES_NAME" validate:"required"`
	SSLMode  string `mapstructure:"POSTGRES_SSLMODE"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST" validate:"required"`
	Port     int    `mapstructure:"REDIS_PORT" validate:"required"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

func Load() (*Config, error) {
	viper.SetDefault("POSTGRES_SSLMODE", "disable")

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &cfg, nil
}
