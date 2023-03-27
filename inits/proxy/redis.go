package proxy

import (
	"github.com/lfxnxf/emo-frame/inits"
	"github.com/lfxnxf/emo-frame/resource/redis"
)

type Redis struct {
	*redis.Redis
}

func InitRedis(name string) *Redis {
	return &Redis{inits.RedisClient(name)}
}
