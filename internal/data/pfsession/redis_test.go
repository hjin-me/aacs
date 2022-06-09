package pfsession

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lunzi/aacs/internal/conf"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ro := NewRedisConf(&conf.Data{Redis: defaultDBConf()})
	r := NewRedis(ro)
	err := r.Set(ctx, "111", "222", 10*time.Second).Err()
	require.NoError(t, err)
	str, err := r.Get(ctx, "111").Result()
	require.NoError(t, err)
	assert.Equal(t, "222", str)
}

func defaultDBConf() *conf.Data_Redis {

	viper.SetDefault("REDIS_DSN", "redis://127.0.0.1:6379/0")
	viper.AutomaticEnv()
	dsn := viper.GetString("REDIS_DSN")
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}
	return &conf.Data_Redis{
		Addr:    opt.Addr,
		Pwd:     opt.Password,
		DbIndex: int32(opt.DB),
	}

}
