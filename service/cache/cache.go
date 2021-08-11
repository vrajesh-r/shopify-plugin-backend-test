package cache

import (
	"errors"

	"github.com/garyburd/redigo/redis"
)

var ErrCacheMiss = errors.New("key does not exist in cache")

type Cache interface {
	Write(key string, value interface{}) (interface{}, error)
	Get(key string) (interface{}, error)
}

type RedisCache struct {
	pool *redis.Pool
}

func NewCache(pool *redis.Pool) Cache {
	return &RedisCache{pool: pool}
}

func (redisCache *RedisCache) Write(key string, value interface{}) (interface{}, error) {
	conn := redisCache.pool.Get()
	defer conn.Close()

	set, err := conn.Do("SET", key, value)
	if err != nil {
		return nil, err
	}

	return set, nil
}

func (redisCache *RedisCache) Get(key string) (interface{}, error) {
	conn := redisCache.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("GET", key)
	if err != nil {
		return nil, err
	}

	if reply == nil {
		return nil, ErrCacheMiss
	}

	return reply, nil
}
