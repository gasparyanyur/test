package config

import (
	"context"

	configLib "node-test/pkg/config"
)

type Config struct {
	Server      ServerConfig  `validate:"required"`
	FileStorage StorageConfig `validate:"required"`
}

type StorageConfig struct {
	Nodes       []string `validate:"required"`
	WorkerCount int      `validate:"required"`
}

type ServerConfig struct {
	Port int `validate:"required,min=80"`
}

func GetConfig(ctx context.Context) (*Config, error) {

	var cfg Config
	err := configLib.Parse(
		ctx,
		configLib.Options{
			Dir:  "./config",
			File: "config.master.yaml",
			Type: "yaml",
		},
		&cfg,
	)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
