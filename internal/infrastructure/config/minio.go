package config

type Minio struct {
	Endpoint   string `yaml:"endpoint" env:"MINIO_ENDPOINT"`
	AccessKey  string `yaml:"access_key" env:"MINIO_ACCESS_KEY"`
	SecretKey  string `yaml:"secret_key" env:"MINIO_SECRET_KEY"`
	UseSSL     bool   `yaml:"use_ssl" env:"MINIO_USE_SSL"`
	BucketName string `yaml:"bucket_name" env:"MINIO_BUCKET_NAME"`
	Region     string `yaml:"region" env:"MINIO_REGION"`
}
