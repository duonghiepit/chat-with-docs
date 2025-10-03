package storage

import (
    "context"
    "time"
    "errors"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
    Pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, url string) (*Database, error) {
    cfg, err := pgxpool.ParseConfig(url)
    if err != nil { return nil, err }
    cfg.MaxConns = 10
    cfg.MaxConnLifetime = time.Hour

    var pool *pgxpool.Pool
    var lastErr error
    deadline := time.Now().Add(45 * time.Second)
    for attempt := 1; time.Now().Before(deadline); attempt++ {
        pool, lastErr = pgxpool.NewWithConfig(ctx, cfg)
        if lastErr == nil {
            if err := pingOnce(ctx, pool); err == nil {
                return &Database{Pool: pool}, nil
            } else {
                lastErr = err
                pool.Close()
            }
        }
        time.Sleep(1500 * time.Millisecond)
    }
    if lastErr == nil { lastErr = errors.New("database connection timeout") }
    return nil, lastErr
}

func pingOnce(ctx context.Context, pool *pgxpool.Pool) error {
    ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    return pool.Ping(ctx2)
}


