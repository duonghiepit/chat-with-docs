package httpserver

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    "strconv"

    "github.com/hiepdt/contest/services/api/internal/llm"
    "github.com/hiepdt/contest/services/api/internal/retrieval"
    "github.com/hiepdt/contest/services/api/internal/storage"
)

type IngestDeps struct {
    Repo *storage.Repository
    LLM  *llm.OllamaClient
    EmbedModel string
    Faiss *retrieval.FaissClient
}

func MakeIngestHandler(deps IngestDeps) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req IngestRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        if req.DocumentID == "" || len(req.Chunks) == 0 {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
        defer cancel()
        if err := deps.Repo.UpsertDocument(ctx, req.DocumentID, ""); err != nil { w.WriteHeader(500); return }
        embeds, err := deps.LLM.Embeddings(ctx, deps.EmbedModel, req.Chunks)
        if err != nil { w.WriteHeader(500); return }
        items := make(map[int64][]float32)
        for i, ch := range req.Chunks {
            var vec []float32
            if i < len(embeds) { vec = embeds[i] }
            // Insert DB row to get id for FAISS
            _ = deps.Repo.InsertChunk(ctx, req.DocumentID, 0, "", ch, vec)
            // naive: use i as temp id; in real impl, return id from DB
            items[int64(i+1)] = vec
        }
        if deps.Faiss != nil { _ = deps.Faiss.Add(ctx, items) }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ingested","chunks":` + strconv.Itoa(len(req.Chunks)) + `}`))
    }
}



