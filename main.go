package main

import "log"

func main() {
	cfg, err := LoadConfigFromFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	store, err := NewMySQLStore(cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}

	redisStore, err := NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.HeartbeatTimeout)
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer(cfg, store, redisStore)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
