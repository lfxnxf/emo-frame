package redis

import (
	"crypto/tls"
	"github.com/lfxnxf/emo-frame/tools/syncx"
	"io"
	"time"

	red "github.com/go-redis/redis/v8"
)

const (
	defaultDatabase = 0
	maxRetries      = 3
	idleConns       = 8
)

var clientManager = syncx.NewResourceManager()

func getClient(r *Redis) (red.Cmdable, error) {
	val, err := clientManager.GetResource(r.Name, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := red.NewClient(&red.Options{
			Addr:         r.Host,
			Password:     r.Pass,
			DB:           r.Database,
			MaxRetries:   r.MaxRetries,
			DialTimeout:  time.Duration(r.DialTimeout) * time.Millisecond,
			ReadTimeout:  time.Duration(r.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(r.WriteTimeout) * time.Millisecond,
			MinIdleConns: r.MinIdle,
			PoolTimeout:  time.Duration(r.PoolTimeout) * time.Millisecond,
			IdleTimeout:  time.Duration(r.IdleTimeout) * time.Millisecond,
			PoolSize:     r.PoolSize,
			TLSConfig:    tlsConfig,
		})
		return store, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*red.Client), nil
}
