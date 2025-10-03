package storage

import (
    "context"
)

const schema = `
CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    title TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chunks (
    id BIGSERIAL PRIMARY KEY,
    document_id TEXT REFERENCES documents(id) ON DELETE CASCADE,
    page INTEGER,
    span TEXT,
    content TEXT NOT NULL,
    embedding VECTOR(768)
);

CREATE TABLE IF NOT EXISTS audits (
    id BIGSERIAL PRIMARY KEY,
    endpoint TEXT,
    latency_ms INTEGER,
    prompt_tokens INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);
`

// RunMigrations creates tables; VECTOR type requires pgvector extension.
// We enable it if available; on vanilla Postgres it's optional (embedding can be NULL).
func (d *Database) RunMigrations(ctx context.Context) error {
    _, err := d.Pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS vector;`)
    if err != nil { /* ignore on systems without superuser */ }
    _, err = d.Pool.Exec(ctx, schema)
    return err
}


