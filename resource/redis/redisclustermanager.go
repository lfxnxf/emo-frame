package redis

import (
	"crypto/tls"
	"github.com/lfxnxf/emo-frame/tools/syncx"
	"io"
	"time"

	red "github.com/go-redis/redis/v8"
)

var clusterManager = syncx.NewResourceManager()

func getCluster(r *Redis) (red.Cmdable, error) {
	val, err := clusterManager.GetResource(r.Host, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        r.AddrList,
			Password:     r.Pass,
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

	return val.(*red.ClusterClient), nil
}
