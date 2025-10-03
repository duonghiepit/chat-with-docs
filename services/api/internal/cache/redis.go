package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type Cache struct {
    Client *redis.Client
}

func New(addr string, db int) *Cache {
    return &Cache{Client: redis.NewClient(&redis.Options{Addr: addr, DB: db})}
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    return c.Client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
    return c.Client.Get(ctx, key).Bytes()
}


