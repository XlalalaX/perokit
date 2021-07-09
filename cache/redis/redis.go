package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/peroperogames/perokit/core/mapping"
	"strconv"
	"time"

	red "github.com/go-redis/redis"
)

const (
	// ClusterType means redis cluster.
	ClusterType = "cluster"
	// NodeType means redis node.
	NodeType = "node"

	PoolType = "pool"
	// Nil is an alias of redis.Nil.
	Nil = red.Nil

	blockingQueryTimeout = 5 * time.Second
	readWriteTimeout     = 2 * time.Second

	slowThreshold = time.Millisecond * 100
	// Minute Time of Ex
	Minute = 60
	Hour   = 60 * Minute
	Day    = 24 * Hour
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
		Addr string
		Type string
		Pass string
		tls  bool
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
)

// New returns a Redis with given options.
func New(addr string, opts ...Option) *Redis {
	r := &Redis{
		Addr: addr,
		Type: NodeType,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// NewRedis returns a Redis.
func NewRedis(redisAddr, redisType string, redisPass ...string) *Redis {
	var opts []Option
	if redisType == ClusterType {
		opts = append(opts, Cluster())
	}
	for _, v := range redisPass {
		opts = append(opts, WithPass(v))
	}

	return New(redisAddr, opts...)
}

// BitCount is redis bitcount command implementation.
func (s *Redis) BitCount(key string, start, end int64) (val int64, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.BitCount(key, &red.BitCount{
		Start: start,
		End:   end,
	}).Result()
	return
}

// BitOpAnd is redis bit operation (and) command implementation.
func (s *Redis) BitOpAnd(destKey string, keys ...string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.BitOpAnd(destKey, keys...).Result()
	return
}

// BitOpNot is redis bit operation (not) command implementation.
func (s *Redis) BitOpNot(destKey, key string) (val int64, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.BitOpNot(destKey, key).Result()
	return
}

// BitOpOr is redis bit operation (or) command implementation.
func (s *Redis) BitOpOr(destKey string, keys ...string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.BitOpOr(destKey, keys...).Result()
	return
}

// BitOpXor is redis bit operation (xor) command implementation.
func (s *Redis) BitOpXor(destKey string, keys ...string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.BitOpXor(destKey, keys...).Result()
	return
}

// BitPos is redis bitpos command implementation.
func (s *Redis) BitPos(key string, bit int64, start, end int64) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.BitPos(key, bit, start, end).Result()
	return
}

// Blpop uses passed in redis connection to execute blocking queries.
// Doesn't benefit from pooling redis connections of blocking queries
func (s *Redis) Blpop(redisNode RedisNode, key string) (string, error) {
	if redisNode == nil {
		return "", ErrNilNode
	}

	vals, err := redisNode.BLPop(blockingQueryTimeout, key).Result()
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
func (s *Redis) BlpopEx(redisNode RedisNode, key string) (string, bool, error) {
	if redisNode == nil {
		return "", false, ErrNilNode
	}

	vals, err := redisNode.BLPop(blockingQueryTimeout, key).Result()
	if err != nil {
		return "", false, err
	}

	if len(vals) < 2 {
		return "", false, fmt.Errorf("no value on key: %s", key)
	}

	return vals[1], true, nil
}

// Del deletes keys.
func (s *Redis) Del(keys ...string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.Del(keys...).Result()
	if err != nil {
		return
	}
	val = int(v)
	return
}

// LikeDel deletes keys.
func (s *Redis) LikeDel(key string) (err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	keys, _ := conn.Keys("*" + key + "*").Result()
	for _, k := range keys {
		_, err = conn.Del(k).Result()
		if err != nil {
			return err
		}
	}
	return
}

// Eval is the implementation of redis eval command.
func (s *Redis) Eval(script string, keys []string, args ...interface{}) (val interface{}, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	val, err = conn.Eval(script, keys, args...).Result()
	return

}

// EvalSha is the implementation of redis evalsha command.
func (s *Redis) EvalSha(sha string, keys []string, args ...interface{}) (val interface{}, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.EvalSha(sha, keys, args...).Result()
	return

}

// Exists is the implementation of redis exists command.
func (s *Redis) Exists(key string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.Exists(key).Result()
	if err != nil {
		return
	}

	val = v == 1
	return
}

// Expire is the implementation of redis expire command.
func (s *Redis) Expire(key string, seconds int) error {

	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Expire(key, time.Duration(seconds)*time.Second).Err()

}

// Expireat is the implementation of redis expireat command.
func (s *Redis) Expireat(key string, expireTime int64) error {

	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.ExpireAt(key, time.Unix(expireTime, 0)).Err()

}

// GeoAdd is the implementation of redis geoadd command.
func (s *Redis) GeoAdd(key string, geoLocation ...*GeoLocation) (val int64, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.GeoAdd(key, geoLocation...).Result()
	if err != nil {
		return
	}

	val = v
	return

}

// GeoDist is the implementation of redis geodist command.
func (s *Redis) GeoDist(key string, member1, member2, unit string) (val float64, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.GeoDist(key, member1, member2, unit).Result()
	if err != nil {
		return
	}

	val = v
	return
}

// GeoHash is the implementation of redis geohash command.
func (s *Redis) GeoHash(key string, members ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.GeoHash(key, members...).Result()
	if err != nil {
		return
	}

	val = v
	return
}

// GeoRadius is the implementation of redis georadius command.
func (s *Redis) GeoRadius(key string, longitude, latitude float64, query *GeoRadiusQuery) (val []GeoLocation, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.GeoRadius(key, longitude, latitude, query).Result()
	if err != nil {
		return
	}

	val = v
	return
}

// GeoRadiusByMember is the implementation of redis georadiusbymember command.
func (s *Redis) GeoRadiusByMember(key, member string, query *GeoRadiusQuery) (val []GeoLocation, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.GeoRadiusByMember(key, member, query).Result()
	if err != nil {
		return
	}

	val = v
	return
}

// GeoPos is the implementation of redis geopos command.
func (s *Redis) GeoPos(key string, members ...string) (val []*GeoPos, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.GeoPos(key, members...).Result()
	if err != nil {
		return
	}

	val = v
	return
}

// Get is the implementation of redis get command.
func (s *Redis) Get(key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	if val, err = conn.Get(key).Result(); err == red.Nil {
		return
	} else if err != nil {
		return
	} else {
		return
	}
}

// GetBit is the implementation of redis getbit command.
func (s *Redis) GetBit(key string, offset int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.GetBit(key, offset).Result()
	if err != nil {
		return
	}

	val = int(v)
	return
}

// Hdel is the implementation of redis hdel command.
func (s *Redis) Hdel(key string, fields ...string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.HDel(key, fields...).Result()
	if err != nil {
		return
	}

	val = v == 1
	return
}

// Hexists is the implementation of redis hexists command.
func (s *Redis) Hexists(key, field string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.HExists(key, field).Result()
	return

}

// Hget is the implementation of redis hget command.
func (s *Redis) Hget(key, field string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.HGet(key, field).Result()
	return

}

// Hgetall is the implementation of redis hgetall command.
func (s *Redis) Hgetall(key string) (val map[string]string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.HGetAll(key).Result()
	return

}

// Hincrby is the implementation of redis hincrby command.
func (s *Redis) Hincrby(key, field string, increment int) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.HIncrBy(key, field, int64(increment)).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Hkeys is the implementation of redis hkeys command.
func (s *Redis) Hkeys(key string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.HKeys(key).Result()
	return

}

// Hlen is the implementation of redis hlen command.
func (s *Redis) Hlen(key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.HLen(key).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Hmget is the implementation of redis hmget command.
func (s *Redis) Hmget(key string, fields ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.HMGet(key, fields...).Result()
	if err != nil {
		return
	}

	val = toStrings(v)
	return
}

// Hset is the implementation of redis hset command.
func (s *Redis) Hset(key, field, value string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.HSet(key, field, value).Err()

}

// Hsetnx is the implementation of redis hsetnx command.
func (s *Redis) Hsetnx(key, field, value string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.HSetNX(key, field, value).Result()
	return

}

// Hmset is the implementation of redis hmset command.
func (s *Redis) Hmset(key string, fieldsAndValues map[string]string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	vals := make(map[string]interface{}, len(fieldsAndValues))
	for k, v := range fieldsAndValues {
		vals[k] = v
	}

	return conn.HMSet(key, vals).Err()
}

// Hscan is the implementation of redis hscan command.
func (s *Redis) Hscan(key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	keys, cur, err = conn.HScan(key, cursor, match, count).Result()
	return

}

// Hvals is the implementation of redis hvals command.
func (s *Redis) Hvals(key string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.HVals(key).Result()
	return
}

// Incr is the implementation of redis incr command.
func (s *Redis) Incr(key string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.Incr(key).Result()
	return
}

// Incrby is the implementation of redis incrby command.
func (s *Redis) Incrby(key string, increment int64) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.IncrBy(key, int64(increment)).Result()
	return
}

// Keys is the implementation of redis keys command.
func (s *Redis) Keys(pattern string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.Keys(pattern).Result()
	return

}

// Llen is the implementation of redis llen command.
func (s *Redis) Llen(key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.LLen(key).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Lpop is the implementation of redis lpop command.
func (s *Redis) Lpop(key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.LPop(key).Result()
	return

}

// Lpush is the implementation of redis lpush command.
func (s *Redis) Lpush(key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.LPush(key, values...).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Lrange is the implementation of redis lrange command.
func (s *Redis) Lrange(key string, start int, stop int) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.LRange(key, int64(start), int64(stop)).Result()
	return

}

// Lrem is the implementation of redis lrem command.
func (s *Redis) Lrem(key string, count int, value string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.LRem(key, int64(count), value).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Mget is the implementation of redis mget command.
func (s *Redis) Mget(keys ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.MGet(keys...).Result()
	if err != nil {
		return
	}

	val = toStrings(v)
	return

}

// Persist is the implementation of redis persist command.
func (s *Redis) Persist(key string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.Persist(key).Result()
	return

}

// Pfadd is the implementation of redis pfadd command.
func (s *Redis) Pfadd(key string, values ...interface{}) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.PFAdd(key, values...).Result()
	if err != nil {
		return
	}

	val = v == 1
	return

}

// Pfcount is the implementation of redis pfcount command.
func (s *Redis) Pfcount(key string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.PFCount(key).Result()
	return

}

// Pfmerge is the implementation of redis pfmerge command.
func (s *Redis) Pfmerge(dest string, keys ...string) error {

	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	_, err = conn.PFMerge(dest, keys...).Result()
	return err

}

// Ping is the implementation of redis ping command.
func (s *Redis) Ping() (val bool) {

	conn, err := getRedis(s)
	if err != nil {
		val = false
		return
	}

	v, err := conn.Ping().Result()
	if err != nil {
		val = false
		return
	}
	val = v == "PONG"
	return

}

// Pipelined lets fn to execute pipelined commands.
func (s *Redis) Pipelined(fn func(Pipeliner) error) (err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	_, err = conn.Pipelined(fn)
	return

}

// Rpop is the implementation of redis rpop command.
func (s *Redis) Rpop(key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.RPop(key).Result()
	return

}

// Rpush is the implementation of redis rpush command.
func (s *Redis) Rpush(key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.RPush(key, values...).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Sadd is the implementation of redis sadd command.
func (s *Redis) Sadd(key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.SAdd(key, values...).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Scan is the implementation of redis scan command.
func (s *Redis) Scan(cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	keys, cur, err = conn.Scan(cursor, match, count).Result()
	return

}

// SetBit is the implementation of redis setbit command.
func (s *Redis) SetBit(key string, offset int64, value int) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	_, err = conn.SetBit(key, offset, value).Result()
	return err

}

// Sscan is the implementation of redis sscan command.
func (s *Redis) Sscan(key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	keys, cur, err = conn.SScan(key, cursor, match, count).Result()
	return

}

// Scard is the implementation of redis scard command.
func (s *Redis) Scard(key string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SCard(key).Result()
	return

}

// ScriptLoad is the implementation of redis script load command.
func (s *Redis) ScriptLoad(script string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	return conn.ScriptLoad(script).Result()
}

// SetStruct  is the implementation of redis set command fro Struct to json.
func (s *Redis) SetStruct(key string, data interface{}) error {
	value, err := json.Marshal(data)
	if err != nil {
		return err
	}
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Set(key, value, 0).Err()

}

// SetStructEx is the implementation of redis setex command fro Struct to json.
func (s *Redis) SetStructEx(key string, data interface{}, seconds int) error {
	value, err := json.Marshal(data)
	if err != nil {
		return err
	}
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Set(key, value, time.Duration(seconds)*time.Second).Err()
}

// Set is the implementation of redis set command.
func (s *Redis) Set(key string, value string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Set(key, value, 0).Err()

}

// Setex is the implementation of redis setex command.
func (s *Redis) Setex(key, value string, seconds int) error {

	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Set(key, value, time.Duration(seconds)*time.Second).Err()
}

// Setnx is the implementation of redis setnx command.
func (s *Redis) Setnx(key, value string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SetNX(key, value, 0).Result()
	return

}

// SetnxEx is the implementation of redis setnx command with expire.
func (s *Redis) SetnxEx(key, value string, seconds int) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SetNX(key, value, time.Duration(seconds)*time.Second).Result()
	return

}

// Sismember is the implementation of redis sismember command.
func (s *Redis) Sismember(key string, value interface{}) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SIsMember(key, value).Result()
	return

}

// Smembers is the implementation of redis smembers command.
func (s *Redis) Smembers(key string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SMembers(key).Result()
	return

}

// Spop is the implementation of redis spop command.
func (s *Redis) Spop(key string) (val string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SPop(key).Result()
	return

}

// Srandmember is the implementation of redis srandmember command.
func (s *Redis) Srandmember(key string, count int) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SRandMemberN(key, int64(count)).Result()
	return

}

// Srem is the implementation of redis srem command.
func (s *Redis) Srem(key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.SRem(key, values...).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// String returns the string representation of s.
func (s *Redis) String() string {
	return s.Addr
}

// Sunion is the implementation of redis sunion command.
func (s *Redis) Sunion(keys ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SUnion(keys...).Result()
	return

}

// Sunionstore is the implementation of redis sunionstore command.
func (s *Redis) Sunionstore(destination string, keys ...string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.SUnionStore(destination, keys...).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Sdiff is the implementation of redis sdiff command.
func (s *Redis) Sdiff(keys ...string) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.SDiff(keys...).Result()
	return

}

// Sdiffstore is the implementation of redis sdiffstore command.
func (s *Redis) Sdiffstore(destination string, keys ...string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.SDiffStore(destination, keys...).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Ttl is the implementation of redis ttl command.
func (s *Redis) Ttl(key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	duration, err := conn.TTL(key).Result()
	if err != nil {
		return
	}

	val = int(duration / time.Second)
	return

}

// Zadd is the implementation of redis zadd command.
func (s *Redis) Zadd(key string, score int64, value string) (val bool, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZAdd(key, red.Z{
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
func (s *Redis) Zadds(key string, ps ...Pair) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	var zs []red.Z
	for _, p := range ps {
		z := red.Z{Score: float64(p.Score), Member: p.Key}
		zs = append(zs, z)
	}

	v, err := conn.ZAdd(key, zs...).Result()
	if err != nil {
		return
	}

	val = v
	return

}

// Zcard is the implementation of redis zcard command.
func (s *Redis) Zcard(key string) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZCard(key).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Zcount is the implementation of redis zcount command.
func (s *Redis) Zcount(key string, start, stop int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZCount(key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Zincrby is the implementation of redis zincrby command.
func (s *Redis) Zincrby(key string, increment int64, field string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZIncrBy(key, float64(increment), field).Result()
	if err != nil {
		return
	}

	val = int64(v)
	return

}

// Zscore is the implementation of redis zscore command.
func (s *Redis) Zscore(key string, value string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZScore(key, value).Result()
	if err != nil {
		return
	}

	val = int64(v)
	return

}

// Zrank is the implementation of redis zrank command.
func (s *Redis) Zrank(key, field string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.ZRank(key, field).Result()
	return

}

// Zrem is the implementation of redis zrem command.
func (s *Redis) Zrem(key string, values ...interface{}) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRem(key, values...).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Zremrangebyscore is the implementation of redis zremrangebyscore command.
func (s *Redis) Zremrangebyscore(key string, start, stop int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRemRangeByScore(key, strconv.FormatInt(start, 10),
		strconv.FormatInt(stop, 10)).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Zremrangebyrank is the implementation of redis zremrangebyrank command.
func (s *Redis) Zremrangebyrank(key string, start, stop int64) (val int, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRemRangeByRank(key, start, stop).Result()
	if err != nil {
		return
	}

	val = int(v)
	return

}

// Zrange is the implementation of redis zrange command.
func (s *Redis) Zrange(key string, start, stop int64) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.ZRange(key, start, stop).Result()
	return

}

// ZrangeWithScores is the implementation of redis zrange command with scores.
func (s *Redis) ZrangeWithScores(key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRangeWithScores(key, start, stop).Result()
	if err != nil {
		return
	}

	val = toPairs(v)
	return

}

// ZRevRangeWithScores is the implementation of redis zrevrange command with scores.
func (s *Redis) ZRevRangeWithScores(key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRevRangeWithScores(key, start, stop).Result()
	if err != nil {
		return
	}

	val = toPairs(v)
	return

}

// ZrangebyscoreWithScores is the implementation of redis zrangebyscore command with scores.
func (s *Redis) ZrangebyscoreWithScores(key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRangeByScoreWithScores(key, red.ZRangeBy{
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
func (s *Redis) ZrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	val []Pair, err error) {
	if size <= 0 {
		return
	}

	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRangeByScoreWithScores(key, red.ZRangeBy{
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
func (s *Redis) Zrevrange(key string, start, stop int64) (val []string, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.ZRevRange(key, start, stop).Result()
	return

}

// ZrevrangebyscoreWithScores is the implementation of redis zrevrangebyscore command with scores.
func (s *Redis) ZrevrangebyscoreWithScores(key string, start, stop int64) (val []Pair, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	v, err := conn.ZRevRangeByScoreWithScores(key, red.ZRangeBy{
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
func (s *Redis) ZrevrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	val []Pair, err error) {
	if size <= 0 {
		return
	}
	conn, err := getRedis(s)
	if err != nil {
		return
	}
	v, err := conn.ZRevRangeByScoreWithScores(key, red.ZRangeBy{
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
func (s *Redis) Zrevrank(key string, field string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}
	val, err = conn.ZRevRank(key, field).Result()
	if err != nil {
		return 0, err
	}
	return val, err
}

// Zunionstore is the implementation of redis zunionstore command.
func (s *Redis) Zunionstore(dest string, store ZStore, keys ...string) (val int64, err error) {
	conn, err := getRedis(s)
	if err != nil {
		return
	}

	val, err = conn.ZUnionStore(dest, store, keys...).Result()
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

func getRedis(r *Redis) (RedisNode, error) {
	switch r.Type {
	case ClusterType:
		return getCluster(r)
	case NodeType:
		return getClient(r)
	case PoolType:
		return getRedisPool(r)
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
