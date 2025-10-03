package storage

import (
    "context"
)

type Repository struct { DB *Database }

func NewRepository(db *Database) *Repository { return &Repository{DB: db} }

func (r *Repository) UpsertDocument(ctx context.Context, id, title string) error {
    _, err := r.DB.Pool.Exec(ctx, `INSERT INTO documents(id, title) VALUES($1,$2)
        ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title`, id, title)
    return err
}

func (r *Repository) InsertChunk(ctx context.Context, docID string, page int, span, content string, embedding []float32) error {
    // Để tránh lỗi kiểu với pgvector khi client chưa đăng ký type, tạm set embedding = NULL.
    _, err := r.DB.Pool.Exec(ctx, `INSERT INTO chunks(document_id,page,span,content,embedding)
        VALUES($1,$2,$3,$4,NULL)`, docID, page, span, content)
    return err
}

// TODO: Add method to return inserted id; simplified for now.

func (r *Repository) SimilarChunks(ctx context.Context, query []float32, topK int) ([]struct{ID int64; DocID string; Content string; Score float32}, error) {
    rows, err := r.DB.Pool.Query(ctx, `
        SELECT id, document_id, content, 1 - (embedding <#> $1) AS score
        FROM chunks WHERE embedding IS NOT NULL
        ORDER BY embedding <-> $1
        LIMIT $2`, query, topK)
    if err != nil { return nil, err }
    defer rows.Close()
    var res []struct{ID int64; DocID string; Content string; Score float32}
    for rows.Next() {
        var it struct{ID int64; DocID string; Content string; Score float32}
        if err := rows.Scan(&it.ID, &it.DocID, &it.Content, &it.Score); err != nil { return nil, err }
        res = append(res, it)
    }
    return res, rows.Err()
}

func (r *Repository) GetChunksByDocument(ctx context.Context, docID string, limit int) ([]string, error) {
    if limit <= 0 { limit = 10 }
    // Lấy các chunk MỚI NHẤT để phản ánh ngữ cảnh vừa ingest
    rows, err := r.DB.Pool.Query(ctx, `SELECT content FROM chunks WHERE document_id=$1 ORDER BY id DESC LIMIT $2`, docID, limit)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []string
    for rows.Next() {
        var c string
        if err := rows.Scan(&c); err != nil { return nil, err }
        out = append(out, c)
    }
    return out, rows.Err()
}

func (r *Repository) SimilarChunksByDoc(ctx context.Context, docID string, query []float32, topK int) ([]struct{ID int64; DocID string; Content string; Score float32}, error) {
    rows, err := r.DB.Pool.Query(ctx, `
        SELECT id, document_id, content, 1 - (embedding <#> $1) AS score
        FROM chunks WHERE document_id=$2 AND embedding IS NOT NULL
        ORDER BY embedding <-> $1
        LIMIT $3`, query, docID, topK)
    if err != nil { return nil, err }
    defer rows.Close()
    var res []struct{ID int64; DocID string; Content string; Score float32}
    for rows.Next() {
        var it struct{ID int64; DocID string; Content string; Score float32}
        if err := rows.Scan(&it.ID, &it.DocID, &it.Content, &it.Score); err != nil { return nil, err }
        res = append(res, it)
    }
    return res, rows.Err()
}


