package httpserver

import (
    "net/http"

    "github.com/go-chi/chi/v5"
)

type API struct{
    // deps are injected from main
    IngestHandler http.HandlerFunc
    SummarizeHandler http.HandlerFunc
    QAHandler http.HandlerFunc
}

func NewRouter(a *API) http.Handler {
    r := chi.NewRouter()
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ok"}`))
    })

    r.Post("/ingest", a.IngestHandler)
    r.Post("/summarize", a.SummarizeHandler)
    r.Post("/qa", a.QAHandler)
    return r
}


