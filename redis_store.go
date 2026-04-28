package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client    *redis.Client
	onlineTTL time.Duration
}

func NewRedisStore(addr string, password string, db int, ttl time.Duration) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStore{client: client, onlineTTL: ttl}, nil
}

func (r *RedisStore) SetOnline(ctx context.Context, user *User) error {
	key := onlineKey(user.Name)
	return r.client.Set(ctx, key, user.Addr, r.onlineTTL).Err()
}

func (r *RedisStore) RefreshOnline(ctx context.Context, user *User) error {
	key := onlineKey(user.Name)
	return r.client.Set(ctx, key, user.Addr, r.onlineTTL).Err()
}

func (r *RedisStore) SetOffline(ctx context.Context, user *User) error {
	key := onlineKey(user.Name)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) RenameUser(ctx context.Context, oldName string, newName string, addr string) error {
	oldKey := onlineKey(oldName)
	newKey := onlineKey(newName)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, oldKey)
	pipe.Set(ctx, newKey, addr, r.onlineTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func onlineKey(name string) string {
	return fmt.Sprintf("im:online:%s", name)
}
