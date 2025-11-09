package config

import (
	"sync"
	"time"
)

type Kafka struct {
	Address         string        `yaml:"address" env:"KAFKA_ADDRESS"`
	ImageTopic      string        `yaml:"image_topic" env:"KAFKA_IMAGE_TOPIC"`
	HealthTopic     string        `yaml:"health_topic" env:"KAFKA_HEALTH_TOPIC"`
	ProcessorGroup  string        `yaml:"processor_group" env:"KAFKA_SAVER_GROUP"`
	CreateTopic     bool          `yaml:"create_topic" env:"KAFKA_CREATE_TOPIC"`
	SessionTimeout  time.Duration `yaml:"session_timeout" env:"KAFKA_SESSION_TIMEOUT" env-default:"30s"`
	MaxPollInterval time.Duration `yaml:"max_poll_interval" env:"KAFKA_MAX_POLL_INTERVAL" env-default:"5m"`
}

var (
	kafkaOnce    sync.Once
	kafkaBrokers []string
)

func (k *Kafka) PrepWbfProducer() ([]string, string) {
	kafkaOnce.Do(func() {
		kafkaBrokers = []string{k.Address}
	})
	return kafkaBrokers, k.ImageTopic
}

func (k *Kafka) PrepWbfConsumer() ([]string, string, string) {
	kafkaOnce.Do(func() {
		kafkaBrokers = []string{k.Address}
	})
	return kafkaBrokers, k.ImageTopic, k.ProcessorGroup
}
