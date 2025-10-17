package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const basicAppConfigPath = "./configs/app/prod.yaml"

type AppConfig struct {
	Storage   Postgres   `yaml:"storage"`
	S3Storage Minio      `yaml:"s3_storage"`
	Broker    Kafka      `yaml:"broker"`
	Server    HTTPServer `yaml:"server"`
}

func NewAppConfig() *AppConfig {
	var cfg AppConfig

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = basicAppConfigPath
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}
