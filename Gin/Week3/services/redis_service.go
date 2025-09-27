package services

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisService(addr, password string, db int) *RedisService {
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	return &RedisService{client: rdb, ctx: context.Background()}
}

func (rs *RedisService) GetClient() *redis.Client {
	return rs.client
}
