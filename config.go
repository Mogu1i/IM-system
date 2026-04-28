package main

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerIp         string        `yaml:"server_ip"`
	ServerPort       int           `yaml:"server_port"`
	WSPath           string        `yaml:"ws_path"`
	MaxConnections   int           `yaml:"max_connections"`
	HeartbeatTimeout time.Duration `yaml:"heartbeat_timeout"`

	PersistBatchSize int           `yaml:"persist_batch_size"`
	PersistInterval  time.Duration `yaml:"persist_interval"`
	PersistQueueSize int           `yaml:"persist_queue_size"`

	MySQLDSN      string `yaml:"mysql_dsn"`
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`
}

func LoadConfigFromFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.MySQLDSN == "" {
		return Config{}, errors.New("mysql_dsn is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, errors.New("redis_addr is required")
	}
	if cfg.MaxConnections <= 0 {
		cfg.MaxConnections = 10000
	}
	if cfg.ServerIp == "" {
		cfg.ServerIp = "127.0.0.1"
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = 8888
	}
	if cfg.WSPath == "" {
		cfg.WSPath = "/ws"
	}
	if cfg.PersistBatchSize <= 0 {
		cfg.PersistBatchSize = 20
	}
	if cfg.PersistInterval <= 0 {
		cfg.PersistInterval = 60 * time.Second
	}
	if cfg.PersistQueueSize <= 0 {
		cfg.PersistQueueSize = 1000
	}
	if cfg.HeartbeatTimeout <= 0 {
		cfg.HeartbeatTimeout = 8 * time.Minute
	}

	return cfg, nil
}
