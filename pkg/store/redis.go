package store

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const DefaultURLTTL = 24 * time.Hour

var ErrClickNotFound = errors.New("redis: click counter not found")

type RedisStore struct {
	cli *redis.Client
}

func NewRedis(addr string) (*RedisStore, error) {
	r := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := r.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStore{cli: r}, nil
}

func (r *RedisStore) Close() error {
	if r.cli == nil {
		return nil
	}
	return r.cli.Close()
}

func (r *RedisStore) Ping(ctx context.Context) error {
	if r.cli == nil {
		return errors.New("redis store: client not initialized")
	}
	return r.cli.Ping(ctx).Err()
}

func (r *RedisStore) key(short string) string      { return "short:" + short }
func (r *RedisStore) clickKey(short string) string { return "click:" + short }

func (r *RedisStore) SetShort(ctx context.Context, short, long string, ttl time.Duration) error {
	if r.cli == nil {
		return errors.New("redis store: client not initialized")
	}
	return r.cli.Set(ctx, r.key(short), long, ttl).Err()
}

func (r *RedisStore) GetShort(ctx context.Context, short string) (string, error) {
	if r.cli == nil {
		return "", errors.New("redis store: client not initialized")
	}
	val, err := r.cli.Get(ctx, r.key(short)).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // cache miss is not an error
	}
	return val, err
}

func (r *RedisStore) DeleteShort(ctx context.Context, short string) error {
	if r.cli == nil {
		return errors.New("redis store: client not initialized")
	}
	return r.cli.Del(ctx, r.key(short)).Err()
}

func (r *RedisStore) IncrClicks(ctx context.Context, short string) (int64, error) {
	if r.cli == nil {
		return 0, errors.New("redis store: client not initialized")
	}
	return r.cli.Incr(ctx, r.clickKey(short)).Result()
}

var getAndClearLua = redis.NewScript(`
local v = redis.call("GET", KEYS[1])
if not v then
  return nil
end
redis.call("DEL", KEYS[1])
return v
`)

// GetAndClearClicks atomically returns the current click counter and clears it.
// It supports different Redis return types and parses them into int64.
func (r *RedisStore) GetAndClearClicks(ctx context.Context, short string) (int64, error) {
	if r.cli == nil {
		return 0, errors.New("redis store: client not initialized")
	}

	key := r.clickKey(short)

	res, err := getAndClearLua.Run(ctx, r.cli, []string{key}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ErrClickNotFound
		}
		return 0, err
	}

	// res can be string, []byte, int64 (not common), or other types depending on client/redis version.
	switch v := res.(type) {
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n, nil
		}
		// try parse as float then convert
		f, err2 := strconv.ParseFloat(v, 64)
		if err2 == nil {
			return int64(f), nil
		}
		return 0, err
	case []byte:
		s := string(v)
		n, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return n, nil
		}
		f, err2 := strconv.ParseFloat(s, 64)
		if err2 == nil {
			return int64(f), nil
		}
		return 0, err
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		// As a last resort, try to stringify via Sprint? but be strict here.
		return 0, errors.New("redis store: unexpected return type from script")
	}
}

// ScanClickKeys can be used by the worker to find active click:* keys.
// Use cursor=0 to start; loop until cursor==0.
func (r *RedisStore) ScanClickKeys(ctx context.Context, cursor uint64, count int64) (uint64, []string, error) {
	if r.cli == nil {
		return 0, nil, errors.New("redis store: client not initialized")
	}

	keys, newCursor, err := r.cli.Scan(ctx, cursor, "click:*", count).Result()
	return newCursor, keys, err
}
