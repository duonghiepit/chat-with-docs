package config

import (
    "os"
)

type Config struct {
    PostgresURL string
    RedisAddr   string
    RedisDB     int
    OllamaHost  string
    ModelName   string
    EmbedModel  string
    FaissHost   string
}

func FromEnv() Config {
    cfg := Config{
        PostgresURL: getenv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/finqa?sslmode=disable"),
        RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
        RedisDB:     0,
        OllamaHost:  getenv("OLLAMA_HOST", "http://localhost:11434"),
        ModelName:   getenv("MODEL_NAME", "qwen2.5:3b"),
        EmbedModel:  getenv("EMBED_MODEL", "bge-m3"),
        FaissHost:   getenv("FAISS_HOST", "http://localhost:8000"),
    }
    return cfg
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}


