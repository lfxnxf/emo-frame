package redis

import (
	"context"
	"errors"
	"fmt"
	red "github.com/go-redis/redis/v8"
	"github.com/lfxnxf/emo-frame/tools/mapping"
	"strconv"
	"time"
)

const (
	// ClusterType means redis cluster.
	ClusterType = "cluster"
	// NodeType means redis node.
	NodeType = "node"
	// Nil is an alias of redis.Nil.
	Nil = red.Nil

	blockingQueryTimeout = 5 * time.Second
	readWriteTimeout     = 2 * time.Second

	slowThreshold = time.Millisecond * 100
)

// ErrNilNode is an error that indicates a nil redis node.
var ErrNilNode = errors.New("nil redis node")

type (
	// Option defines the method to customize a Redis.
	Option func(r *Redis)

	// A Pair is a key/pair set used in redis zset.
	Pair struct {
		Key   string
		Score int64
	}

	// Redis defines a redis node/cluster. It is thread-safe.
	Redis struct {
		Conf
		tls bool
	}

	// RedisNode interface represents a redis node.
	RedisNode interface {
		red.Cmdable
	}

	// GeoLocation is used with GeoAdd to add geospatial location.
	GeoLocation = red.GeoLocation
	// GeoRadiusQuery is used with GeoRadius to query geospatial index.
	GeoRadiusQuery = red.GeoRadiusQuery
	// GeoPos is used to represent a geo position.
	GeoPos = red.GeoPos

	// Pipeliner is an alias of redis.Pipeliner.
	Pipeliner = red.Pipeliner

	// Z represents sorted set member.
	Z = red.Z
	// ZStore is an alias of redis.ZStore.
	ZStore = red.ZStore

	// IntCmd is an alias of redis.IntCmd.
	IntCmd = red.IntCmd
	// FloatCmd is an alias of redis.FloatCmd.
	FloatCmd = red.FloatCmd
	// StringCmd is an alias of redis.StringCmd.
	StringCmd = red.StringCmd
)

// New returns a Redis with given options.
func New(rc Conf, opts ...Option) *Redis {
	r := &Redis{
		Conf: rc,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Deprecated: use New instead, will be removed in v2.
// NewRedis returns a Redis.
func NewRedis(redisAddr, redisType string, redisPass ...string) *Redis {
	var opts []Option
	if redisType == ClusterType {
		opts = append(opts, Cluster())
	}
	for _, v := range redisPass {
		opts = append(opts, WithPass(v))
	}

	return New(Conf{Host: redisAddr}, opts...)
}

// BitCount is redis bitcount command implementation.
func (s *Redis) BitCount(ctx context.Context, key string, start, end int64) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.BitCount(ctx, key, &red.BitCount{
		Start: start,
		End:   end,
	}).Result()
	return
}

// BitOpAnd is redis bit operation (and) command implementation.
func (s *Redis) BitOpAnd(ctx context.Context, destKey string, keys ...string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.BitOpAnd(ctx, destKey, keys...).Result()
	return
}

// BitOpNot is redis bit operation (not) command implementation.
func (s *Redis) BitOpNot(ctx context.Context, destKey, key string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.BitOpNot(ctx, destKey, key).Result()
	return
}

// BitOpOr is redis bit operation (or) command implementation.
func (s *Redis) BitOpOr(ctx context.Context, destKey string, keys ...string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.BitOpOr(ctx, destKey, keys...).Result()
	return
}

// BitOpXor is redis bit operation (xor) command implementation.
func (s *Redis) BitOpXor(ctx context.Context, destKey string, keys ...string) (val int64, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.BitOpXor(ctx, destKey, keys...).Result()
	return
}

// BitPos is redis bitpos command implementation.
func (s *Redis) BitPos(ctx context.Context, key string, bit, start, end int64) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.BitPos(ctx, key, bit, start, end).Result()
	return
}

// Blpop uses passed in redis connection to execute blocking queries.
// Doesn't benefit from pooling redis connections of blocking queries
func (s *Redis) Blpop(ctx context.Context, redisNode RedisNode, key string, timeout time.Duration) (string, error) {
	if redisNode == nil {
		return "", ErrNilNode
	}
	if timeout <= 0 {
		timeout = blockingQueryTimeout
	}
	vals, err := redisNode.BLPop(ctx, timeout, key).Result()
	if err != nil {
		return "", err
	}
	if len(vals) < 2 {
		return "", fmt.Errorf("no value on key: %s", key)
	}
	return vals[1], nil
}

// BlpopEx uses passed in redis connection to execute blpop command.
// The difference against Blpop is that this method returns a bool to indicate success.
func (s *Redis) BlpopEx(ctx context.Context, redisNode RedisNode, key string, timeout time.Duration) (string, bool, error) {
	if redisNode == nil {
		return "", false, ErrNilNode
	}

	if timeout <= 0 {
		timeout = blockingQueryTimeout
	}

	vals, err := redisNode.BLPop(ctx, timeout, key).Result()
	if err != nil {
		return "", false, err
	}

	if len(vals) < 2 {
		return "", false, fmt.Errorf("no value on key: %s", key)
	}

	return vals[1], true, nil
}

// Del deletes keys.
func (s *Redis) Del(ctx context.Context, keys ...string) (val int, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.Del(ctx, keys...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Eval is the implementation of redis eval command.
func (s *Redis) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (val interface{}, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.Eval(ctx, script, keys, args...).Result()
	return
}

// EvalSha is the implementation of redis evalsha command.
func (s *Redis) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) (val interface{}, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.EvalSha(ctx, sha, keys, args...).Result()
	return
}

// Exists is the implementation of redis exists command.
func (s *Redis) Exists(ctx context.Context, key string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.Exists(ctx, key).Result()
	if err != nil {
		return
	}
	val = v == 1
	return
}

// Expire is the implementation of redis expire command.
func (s *Redis) Expire(ctx context.Context, key string, seconds int) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
}

// Expireat is the implementation of redis expireat command.
func (s *Redis) Expireat(ctx context.Context, key string, expireTime int64) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.ExpireAt(ctx, key, time.Unix(expireTime, 0)).Err()
}

// GeoAdd is the implementation of redis geoadd command.
func (s *Redis) GeoAdd(ctx context.Context, key string, geoLocation ...*GeoLocation) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.GeoAdd(ctx, key, geoLocation...).Result()
	if err != nil {
		return
	}
	val = v
	return
}

// GeoDist is the implementation of redis geodist command.
func (s *Redis) GeoDist(ctx context.Context, key, member1, member2, unit string) (val float64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.GeoDist(ctx, key, member1, member2, unit).Result()
	if err != nil {
		return
	}
	val = v
	return
}

// GeoHash is the implementation of redis geohash command.
func (s *Redis) GeoHash(ctx context.Context, key string, members ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.GeoHash(ctx, key, members...).Result()
	if err != nil {
		return
	}
	val = v
	return
}

// GeoRadius is the implementation of redis georadius command.
func (s *Redis) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *GeoRadiusQuery) (val []GeoLocation, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.GeoRadius(ctx, key, longitude, latitude, query).Result()
	if err != nil {
		return
	}
	val = v
	return
}

// GeoRadiusByMember is the implementation of redis georadiusbymember command.
func (s *Redis) GeoRadiusByMember(ctx context.Context, key, member string, query *GeoRadiusQuery) (val []GeoLocation, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.GeoRadiusByMember(ctx, key, member, query).Result()
	if err != nil {
		return
	}
	val = v
	return
}

// GeoPos is the implementation of redis geopos command.
func (s *Redis) GeoPos(ctx context.Context, key string, members ...string) (val []*GeoPos, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.GeoPos(ctx, key, members...).Result()
	if err != nil {
		return
	}
	val = v
	return
}

// Get is the implementation of redis get command.
func (s *Redis) Get(ctx context.Context, key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.Get(ctx, key).Result()

	// 特殊处理redis.Nil
	if err == red.Nil {
		err = nil
		val = ""
	}
	return
}

// GetBit is the implementation of redis getbit command.
func (s *Redis) GetBit(ctx context.Context, key string, offset int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.GetBit(ctx, key, offset).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Hdel is the implementation of redis hdel command.
func (s *Redis) Hdel(ctx context.Context, key string, fields ...string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.HDel(ctx, key, fields...).Result()
	if err != nil {
		return
	}
	val = v == 1
	return
}

// Hexists is the implementation of redis hexists command.
func (s *Redis) Hexists(ctx context.Context, key, field string) (val bool, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.HExists(ctx, key, field).Result()
	return
}

// Hget is the implementation of redis hget command.
func (s *Redis) Hget(ctx context.Context, key, field string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.HGet(ctx, key, field).Result()
	return
}

// Hgetall is the implementation of redis hgetall command.
func (s *Redis) Hgetall(ctx context.Context, key string) (val map[string]string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.HGetAll(ctx, key).Result()
	return
}

// Hincrby is the implementation of redis hincrby command.
func (s *Redis) Hincrby(ctx context.Context, key, field string, increment int) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.HIncrBy(ctx, key, field, int64(increment)).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Hkeys is the implementation of redis hkeys command.
func (s *Redis) Hkeys(ctx context.Context, key string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.HKeys(ctx, key).Result()
	return
}

// Hlen is the implementation of redis hlen command.
func (s *Redis) Hlen(ctx context.Context, key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.HLen(ctx, key).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Hmget is the implementation of redis hmget command.
func (s *Redis) Hmget(ctx context.Context, key string, fields ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return
	}
	val = toStrings(v)
	return
}

// Hset is the implementation of redis hset command.
func (s *Redis) Hset(ctx context.Context, key, field, value string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.HSet(ctx, key, field, value).Err()
}

// Hsetnx is the implementation of redis hsetnx command.
func (s *Redis) Hsetnx(ctx context.Context, key, field, value string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.HSetNX(ctx, key, field, value).Result()
	return
}

// Hmset is the implementation of redis hmset command.
func (s *Redis) Hmset(ctx context.Context, key string, fieldsAndValues map[string]string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	vals := make(map[string]interface{}, len(fieldsAndValues))
	for k, v := range fieldsAndValues {
		vals[k] = v
	}
	return conn.HMSet(ctx, key, vals).Err()
}

// Hscan is the implementation of redis hscan command.
func (s *Redis) Hscan(ctx context.Context, key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	keys, cur, err = conn.HScan(ctx, key, cursor, match, count).Result()
	return
}

// Hvals is the implementation of redis hvals command.
func (s *Redis) Hvals(ctx context.Context, key string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.HVals(ctx, key).Result()
	return
}

// Incr is the implementation of redis incr command.
func (s *Redis) Incr(ctx context.Context, key string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.Incr(ctx, key).Result()
	return
}

// Incrby is the implementation of redis incrby command.
func (s *Redis) Incrby(ctx context.Context, key string, increment int64) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.IncrBy(ctx, key, int64(increment)).Result()
	return
}

// Keys is the implementation of redis keys command.
func (s *Redis) Keys(ctx context.Context, pattern string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.Keys(ctx, pattern).Result()
	return
}

// Llen is the implementation of redis llen command.
func (s *Redis) Llen(ctx context.Context, key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.LLen(ctx, key).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Lpop is the implementation of redis lpop command.
func (s *Redis) Lpop(ctx context.Context, key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.LPop(ctx, key).Result()
	return
}

// Lpush is the implementation of redis lpush command.
func (s *Redis) Lpush(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.LPush(ctx, key, values...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Lrange is the implementation of redis lrange command.
func (s *Redis) Lrange(ctx context.Context, key string, start, stop int) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.LRange(ctx, key, int64(start), int64(stop)).Result()
	return
}

// Lrem is the implementation of redis lrem command.
func (s *Redis) Lrem(ctx context.Context, key string, count int, value string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.LRem(ctx, key, int64(count), value).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Mget is the implementation of redis mget command.
func (s *Redis) Mget(ctx context.Context, keys ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.MGet(ctx, keys...).Result()
	if err != nil {
		return
	}
	val = toStrings(v)
	return
}

// Persist is the implementation of redis persist command.
func (s *Redis) Persist(ctx context.Context, key string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.Persist(ctx, key).Result()
	return
}

// Pfadd is the implementation of redis pfadd command.
func (s *Redis) Pfadd(ctx context.Context, key string, values ...interface{}) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.PFAdd(ctx, key, values...).Result()
	if err != nil {
		return
	}
	val = v == 1
	return
}

// Pfcount is the implementation of redis pfcount command.
func (s *Redis) Pfcount(ctx context.Context, key string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.PFCount(ctx, key).Result()
	return
}

// Pfmerge is the implementation of redis pfmerge command.
func (s *Redis) Pfmerge(ctx context.Context, dest string, keys ...string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	_, err = conn.PFMerge(ctx, dest, keys...).Result()
	return nil
}

// Ping is the implementation of redis ping command.
func (s *Redis) Ping(ctx context.Context) (val bool) {
	conn, err := getRedis(s)
	if err != nil {
		val = false
		return
	}
	v, err := conn.Ping(ctx).Result()
	if err != nil {
		val = false
		return
	}
	val = v == "PONG"
	return
}

// Rpop is the implementation of redis rpop command.
func (s *Redis) Rpop(ctx context.Context, key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.RPop(ctx, key).Result()
	return
}

// Rpush is the implementation of redis rpush command.
func (s *Redis) Rpush(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.RPush(ctx, key, values...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Sadd is the implementation of redis sadd command.
func (s *Redis) Sadd(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.SAdd(ctx, key, values...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Scan is the implementation of redis scan command.
func (s *Redis) Scan(ctx context.Context, cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	keys, cur, err = conn.Scan(ctx, cursor, match, count).Result()
	return
}

// SetBit is the implementation of redis setbit command.
func (s *Redis) SetBit(ctx context.Context, key string, offset int64, value int) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	_, err = conn.SetBit(ctx, key, offset, value).Result()
	return err
}

// Sscan is the implementation of redis sscan command.
func (s *Redis) Sscan(ctx context.Context, key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	keys, cur, err = conn.SScan(ctx, key, cursor, match, count).Result()
	return
}

// Scard is the implementation of redis scard command.
func (s *Redis) Scard(ctx context.Context, key string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SCard(ctx, key).Result()
	return
}

// ScriptLoad is the implementation of redis script load command.
func (s *Redis) ScriptLoad(ctx context.Context, script string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}
	return conn.ScriptLoad(ctx, script).Result()
}

// Set is the implementation of redis set command.
func (s *Redis) Set(ctx context.Context, key string, value interface{}) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.Set(ctx, key, value, 0).Err()
}

// Setex is the implementation of redis setex command.
func (s *Redis) Setex(ctx context.Context, key, value string, seconds int) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.Set(ctx, key, value, time.Duration(seconds)*time.Second).Err()
}

// Setnx is the implementation of redis setnx command.
func (s *Redis) Setnx(ctx context.Context, key, value string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SetNX(ctx, key, value, 0).Result()
	return
}

// SetnxEx is the implementation of redis setnx command with expire.
func (s *Redis) SetnxEx(ctx context.Context, key, value string, seconds int) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SetNX(ctx, key, value, time.Duration(seconds)*time.Second).Result()
	return
}

// Sismember is the implementation of redis sismember command.
func (s *Redis) Sismember(ctx context.Context, key string, value interface{}) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SIsMember(ctx, key, value).Result()
	return
}

// Smembers is the implementation of redis smembers command.
func (s *Redis) Smembers(ctx context.Context, key string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SMembers(ctx, key).Result()
	return
}

// Spop is the implementation of redis spop command.
func (s *Redis) Spop(ctx context.Context, key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SPop(ctx, key).Result()
	return
}

// Srandmember is the implementation of redis srandmember command.
func (s *Redis) Srandmember(ctx context.Context, key string, count int) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SRandMemberN(ctx, key, int64(count)).Result()
	return
}

// Srem is the implementation of redis srem command.
func (s *Redis) Srem(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.SRem(ctx, key, values...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// String returns the string representation of s.
func (s *Redis) String() string {
	return s.Host
}

// Sunion is the implementation of redis sunion command.
func (s *Redis) Sunion(ctx context.Context, keys ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SUnion(ctx, keys...).Result()
	return
}

// Sunionstore is the implementation of redis sunionstore command.
func (s *Redis) Sunionstore(ctx context.Context, destination string, keys ...string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.SUnionStore(ctx, destination, keys...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Sdiff is the implementation of redis sdiff command.
func (s *Redis) Sdiff(ctx context.Context, keys ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SDiff(ctx, keys...).Result()
	return
}

// Sdiffstore is the implementation of redis sdiffstore command.
func (s *Redis) Sdiffstore(ctx context.Context, destination string, keys ...string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.SDiffStore(ctx, destination, keys...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Sinter is the implementation of redis sinter command.
func (s *Redis) Sinter(ctx context.Context, keys ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.SInter(ctx, keys...).Result()
	return
}

// Sinterstore is the implementation of redis sinterstore command.
func (s *Redis) Sinterstore(ctx context.Context, destination string, keys ...string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.SInterStore(ctx, destination, keys...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Ttl is the implementation of redis ttl command.
func (s *Redis) Ttl(ctx context.Context, key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	duration, err := conn.TTL(ctx, key).Result()
	if err != nil {
		return
	}
	val = int(duration / time.Second)
	return
}

// Zadd is the implementation of redis zadd command.
func (s *Redis) Zadd(ctx context.Context, key string, score int64, value string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZAdd(ctx, key, &red.Z{
		Score:  float64(score),
		Member: value,
	}).Result()
	if err != nil {
		return
	}
	val = v == 1
	return
}

// Zadds is the implementation of redis zadds command.
func (s *Redis) Zadds(ctx context.Context, key string, ps ...Pair) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	var zs []*red.Z
	for _, p := range ps {
		z := &red.Z{Score: float64(p.Score), Member: p.Key}
		zs = append(zs, z)
	}
	v, err := conn.ZAdd(ctx, key, zs...).Result()
	if err != nil {
		return
	}
	val = v
	return
}

// Zcard is the implementation of redis zcard command.
func (s *Redis) Zcard(ctx context.Context, key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZCard(ctx, key).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Zcount is the implementation of redis zcount command.
func (s *Redis) Zcount(ctx context.Context, key string, start, stop int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZCount(ctx, key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Zincrby is the implementation of redis zincrby command.
func (s *Redis) Zincrby(ctx context.Context, key string, increment int64, field string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZIncrBy(ctx, key, float64(increment), field).Result()
	if err != nil {
		return
	}
	val = int64(v)
	return
}

// Zscore is the implementation of redis zscore command.
func (s *Redis) Zscore(ctx context.Context, key, value string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZScore(ctx, key, value).Result()
	if err != nil {
		return
	}
	val = int64(v)
	return
}

// Zrank is the implementation of redis zrank command.
func (s *Redis) Zrank(ctx context.Context, key, field string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.ZRank(ctx, key, field).Result()
	return
}

// Zrem is the implementation of redis zrem command.
func (s *Redis) Zrem(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRem(ctx, key, values...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Zremrangebyscore is the implementation of redis zremrangebyscore command.
func (s *Redis) Zremrangebyscore(ctx context.Context, key string, start, stop int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRemRangeByScore(ctx, key, strconv.FormatInt(start, 10),
		strconv.FormatInt(stop, 10)).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Zremrangebyrank is the implementation of redis zremrangebyrank command.
func (s *Redis) Zremrangebyrank(ctx context.Context, key string, start, stop int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRemRangeByRank(ctx, key, start, stop).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// Zrange is the implementation of redis zrange command.
func (s *Redis) Zrange(ctx context.Context, key string, start, stop int64) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.ZRange(ctx, key, start, stop).Result()
	return
}

// ZrangeWithScores is the implementation of redis zrange command with scores.
func (s *Redis) ZrangeWithScores(ctx context.Context, key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return
	}
	val = toPairs(v)
	return
}

// ZRevRangeWithScores is the implementation of redis zrevrange command with scores.
func (s *Redis) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return
	}
	val = toPairs(v)
	return
}

// ZrangebyscoreWithScores is the implementation of redis zrangebyscore command with scores.
func (s *Redis) ZrangebyscoreWithScores(ctx context.Context, key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
		Min: strconv.FormatInt(start, 10),
		Max: strconv.FormatInt(stop, 10),
	}).Result()
	if err != nil {
		return
	}
	val = toPairs(v)
	return
}

// ZrangebyscoreWithScoresAndLimit is the implementation of redis zrangebyscore command with scores and limit.
func (s *Redis) ZrangebyscoreWithScoresAndLimit(ctx context.Context, key string, start, stop int64, page, size int) (
	val []Pair, err error) {
	if size <= 0 {
		return
	}
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
		Min:    strconv.FormatInt(start, 10),
		Max:    strconv.FormatInt(stop, 10),
		Offset: int64(page * size),
		Count:  int64(size),
	}).Result()
	if err != nil {
		return
	}
	val = toPairs(v)
	return
}

// Zrevrange is the implementation of redis zrevrange command.
func (s *Redis) Zrevrange(ctx context.Context, key string, start, stop int64) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.ZRevRange(ctx, key, start, stop).Result()
	return
}

// ZrevrangebyscoreWithScores is the implementation of redis zrevrangebyscore command with scores.
func (s *Redis) ZrevrangebyscoreWithScores(ctx context.Context, key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRevRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
		Min: strconv.FormatInt(start, 10),
		Max: strconv.FormatInt(stop, 10),
	}).Result()
	if err != nil {
		return
	}
	val = toPairs(v)
	return
}

// ZrevrangebyscoreWithScoresAndLimit is the implementation of redis zrevrangebyscore command with scores and limit.
func (s *Redis) ZrevrangebyscoreWithScoresAndLimit(ctx context.Context, key string, start, stop int64, page, size int) (
	val []Pair, err error) {
	if size <= 0 {
		return
	}
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRevRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
		Min:    strconv.FormatInt(start, 10),
		Max:    strconv.FormatInt(stop, 10),
		Offset: int64(page * size),
		Count:  int64(size),
	}).Result()
	if err != nil {
		return
	}
	val = toPairs(v)
	return
}

// Zrevrank is the implementation of redis zrevrank command.
func (s *Redis) Zrevrank(ctx context.Context, key, field string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.ZRevRank(ctx, key, field).Result()
	return
}

// Zunionstore is the implementation of redis zunionstore command.
func (s *Redis) Zunionstore(ctx context.Context, dest string, store *ZStore) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.ZUnionStore(ctx, dest, store).Result()
	return
}

// Cluster customizes the given Redis as a cluster.
func Cluster() Option {
	return func(r *Redis) {
		r.Type = ClusterType
	}
}

// WithPass customizes the given Redis with given password.
func WithPass(pass string) Option {
	return func(r *Redis) {
		r.Pass = pass
	}
}

// WithTLS customizes the given Redis with TLS enabled.
func WithTLS() Option {
	return func(r *Redis) {
		r.tls = true
	}
}

func acceptable(err error) bool {
	return err == nil || err == red.Nil
}

func getRedis(r *Redis) (RedisNode, error) {
	if r.Type == "" {
		r.Type = NodeType
	}
	switch r.Type {
	case ClusterType:
		return getCluster(r)
	case NodeType:
		return getClient(r)
	default:
		return nil, fmt.Errorf("redis type '%s' is not supported", r.Type)
	}
}

func toPairs(vals []red.Z) []Pair {
	pairs := make([]Pair, len(vals))
	for i, val := range vals {
		switch member := val.Member.(type) {
		case string:
			pairs[i] = Pair{
				Key:   member,
				Score: int64(val.Score),
			}
		default:
			pairs[i] = Pair{
				Key:   mapping.Repr(val.Member),
				Score: int64(val.Score),
			}
		}
	}
	return pairs
}

func toStrings(vals []interface{}) []string {
	ret := make([]string, len(vals))
	for i, val := range vals {
		if val == nil {
			ret[i] = ""
		} else {
			switch val := val.(type) {
			case string:
				ret[i] = val
			default:
				ret[i] = mapping.Repr(val)
			}
		}
	}
	return ret
}
