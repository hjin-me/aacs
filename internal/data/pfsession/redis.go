package pfsession

import (
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/lunzi/aacs/internal/conf"
)

func NewRedisConf(data *conf.Data) *redis.Options {
	return &redis.Options{
		Addr:     data.Redis.Addr,
		Password: data.Redis.Pwd,          // no password set
		DB:       int(data.Redis.DbIndex), // use default DB
	}
}

func NewRedis(ro *redis.Options) *redis.Client {
	rdb := redis.NewClient(ro)
	rdb.AddHook(redisotel.NewTracingHook())
	return rdb
}
