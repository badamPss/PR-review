package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type LogConfig struct {
	Level string `yaml:"level"`
	AppID string `yaml:"app_id"`
}

type HTTPServerConfig struct {
	Listen       string        `yaml:"listen"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type ConnConfig struct {
	Network  string `yaml:"network"`
	Database string `yaml:"database"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type SQLConfig struct {
	ConnConfig      `yaml:"conn_config"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnLifeTime    time.Duration `yaml:"conn_life_time"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

type DatabaseConfig struct {
	Postgres SQLConfig `yaml:"postgres"`
}

type RateLimitConfig struct {
	Requests int `yaml:"requests"`
	Burst    int `yaml:"burst"`
}

type Config struct {
	Log             LogConfig        `yaml:"log"`
	HTTPServer      HTTPServerConfig `yaml:"http_server"`
	Database        DatabaseConfig   `yaml:"database"`
	GracefulTimeout time.Duration    `yaml:"graceful_timeout"`
	RateLimit       RateLimitConfig  `yaml:"rate_limit"`
}

func ReadConfig(paths ...string) (*Config, error) {
	var config Config
	for _, path := range paths {
		yamlFile, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		expandedData := os.ExpandEnv(string(yamlFile))

		err = yaml.Unmarshal([]byte(expandedData), &config)
		if err != nil {
			return nil, err
		}
	}

	return &config, nil
}
