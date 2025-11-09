package config

import (
	"time"
)

type HTTPServer struct {
	Host         string        `yaml:"host" env:"HTTP_ADDRESS" env-default:"localhost"`
	Port         string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	Timeout      time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ReleaseMode  bool          `yaml:"release_mode" env:"HTTP_RELEASE_MODE" env-default:"false"`
	CORS         bool          `yaml:"cors" env:"HTTP_CORS"`
	AllowOrigins []string      `yaml:"allow_origins" env:"HTTP_ALLOWED_ORIGINS"`
	ServeUI      bool          `yaml:"serve_ui" env:"HTTP_SERVE_UI" env-default:"false"`
	UIPath       string        `yaml:"ui_path" env:"HTTP_UI_PATH" env-default:"/ui/index.html"`
}
