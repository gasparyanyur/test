package config

import (
	"context"

	configLib "node-test/pkg/config"
	"node-test/pkg/mongodb"
)

type Config struct {
	Server ServerConfig `validate:"required"`
	Mongo  MongoConfig  `validate:"required"`
}

type MongoConfig struct {
	URI      string `validate:"required,url"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	DB       string `validate:"required"`
}

func (cfg MongoConfig) External() mongodb.Config {
	return mongodb.Config{
		URI:      cfg.URI,
		User:     cfg.User,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}

type ServerConfig struct {
	Port int `validate:"required,min=80"`
	// free size for node in bytes
	Size int64 `validate:"required"`
}

func GetConfig(ctx context.Context) (*Config, error) {

	var cfg Config
	err := configLib.Parse(
		ctx,
		configLib.Options{
			Dir:  "./config",
			File: "config.node.yaml",
			Type: "yaml",
		},
		&cfg,
	)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
