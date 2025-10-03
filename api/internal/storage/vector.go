package storage

// Vector is a helper type to marshal []float32 to Postgres pgvector.
// pgx can send []float32 directly if pgvector extension type registered.
// For simplicity, we pass as []float32 and rely on default mapping where available.


