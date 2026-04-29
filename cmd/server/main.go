package main

import (
	"log"

	"IM-System/internal/config"
	"IM-System/internal/redisstore"
	"IM-System/internal/server"
	"IM-System/internal/storage"
)

func main() {
	cfg, err := config.LoadConfigFromFile("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.NewMySQLStore(cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}

	redisStore, err := redisstore.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.HeartbeatTimeout)
	if err != nil {
		log.Fatal(err)
	}

	imServer := server.NewServer(cfg, store, redisStore)
	if err := imServer.Start(); err != nil {
		log.Fatal(err)
	}
}
