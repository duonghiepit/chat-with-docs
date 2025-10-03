package httpserver

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

type IngestRequest struct {
    DocumentID string   `json:"document_id"`
    Chunks     []string `json:"chunks"`
}

type QARequest struct {
    Question string `json:"question"`
    TopK     int    `json:"top_k"`
}

type SummarizeRequest struct {
    DocumentID string `json:"document_id"`
    NumBullets int    `json:"num_bullets"`
    Category   string `json:"category"`
    Instruction string `json:"instruction"`
}

func (a *API) writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func (a *API) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
    return context.WithTimeout(ctx, 60*time.Second)
}


