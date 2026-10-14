package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppPort         string
	AppEnv          string
	MongoURI        string
	JWTAccessSecret string
}

func Load() (Config, error) {
	cfg := Config{
		AppPort:         strings.TrimSpace(os.Getenv("APP_PORT")),
		AppEnv:          strings.TrimSpace(os.Getenv("APP_ENV")),
		MongoURI:        strings.TrimSpace(os.Getenv("MONGODB_URI")),
		JWTAccessSecret: strings.TrimSpace(os.Getenv("JWT_ACCESS_SECRET")),
	}

	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	if cfg.AppEnv != "development" &&
		cfg.AppEnv != "production" &&
		cfg.AppEnv != "test" {
		return Config{}, fmt.Errorf(
			"invalid APP_ENV %q: allowed values are development, production, or test",
			cfg.AppEnv,
		)
	}

	if cfg.MongoURI == "" {
		return Config{}, fmt.Errorf("MONGODB_URI is required")
	}

	if len(cfg.JWTAccessSecret) < 32 {
		return Config{}, fmt.Errorf(
			"JWT_ACCESS_SECRET must contain at least 32 characters",
		)
	}

	return cfg, nil
}
