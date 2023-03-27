package redis

import "errors"

var (
	// ErrEmptyHost is an error that indicates no redis host is set.
	ErrEmptyHost = errors.New("empty redis host")
	// ErrEmptyType is an error that indicates no redis type is set.
	ErrEmptyType = errors.New("empty redis type")
	// ErrEmptyKey is an error that indicates no redis key is set.
	ErrEmptyKey = errors.New("empty redis key")
)

type (
	// A RedisConf is a redis config.
	Conf struct {
		Name         string   `toml:"name"`
		Host         string   `toml:"host"`      // node时传
		AddrList     []string `toml:"addr_list"` // cluster时传
		Pass         string   `toml:"pass"`
		MinIdle      int      `toml:"min_idle"`
		Database     int      `toml:"database"`
		MaxRetries   int      `toml:"max_retries"`
		DialTimeout  int      `toml:"dial_timeout"`
		ReadTimeout  int      `toml:"read_timeout"`
		WriteTimeout int      `toml:"write_timeout"`
		PoolSize     int      `toml:"pool_size"`
		PoolTimeout  int      `toml:"pool_timeout"`
		IdleTimeout  int      `toml:"idle_timeout"`
		Type         string   `toml:"type"`
		Tls          bool     `toml:"tls"`
	}

	// A RedisKeyConf is a redis config with key.
	RedisKeyConf struct {
		Conf
		Key string `json:",optional"`
	}
)

// NewRedis returns a Redis.
func (rc Conf) NewRedis() *Redis {
	var opts []Option
	if rc.Type == ClusterType {
		opts = append(opts, Cluster())
	}
	if len(rc.Pass) > 0 {
		opts = append(opts, WithPass(rc.Pass))
	}
	if rc.Tls {
		opts = append(opts, WithTLS())
	}
	return New(rc, opts...)
}

// Validate validates the RedisConf.
func (rc Conf) Validate() error {
	if len(rc.Host) == 0 {
		return ErrEmptyHost
	}

	if len(rc.Type) == 0 {
		return ErrEmptyType
	}

	return nil
}

// Validate validates the RedisKeyConf.
func (rkc RedisKeyConf) Validate() error {
	if err := rkc.Conf.Validate(); err != nil {
		return err
	}

	if len(rkc.Key) == 0 {
		return ErrEmptyKey
	}

	return nil
}
