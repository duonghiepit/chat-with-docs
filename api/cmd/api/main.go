package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"

    "github.com/hiepdt/contest/services/api/internal/cache"
    "github.com/hiepdt/contest/services/api/internal/config"
    "github.com/hiepdt/contest/services/api/internal/llm"
    "github.com/hiepdt/contest/services/api/internal/httpserver"
    "github.com/hiepdt/contest/services/api/internal/retrieval"
    "github.com/hiepdt/contest/services/api/internal/storage"
    "github.com/hiepdt/contest/services/api/internal/metrics"
)

func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}

func run() error {
    cfg := config.FromEnv()

    // setup router
    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "*"},
        AllowedMethods:   []string{"GET","POST","PUT","DELETE","OPTIONS"},
        AllowedHeaders:   []string{"*"},
        ExposedHeaders:   []string{"*"},
        AllowCredentials: false,
    }))

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ok"}`))
    })
    r.Handle("/metrics", metrics.Handler())

    // init deps
    ctx := context.Background()
    db, err := storage.NewDatabase(ctx, cfg.PostgresURL)
    if err != nil { return err }
    if err := db.RunMigrations(ctx); err != nil { log.Println("migrate:", err) }
    _ = cache.New(cfg.RedisAddr, cfg.RedisDB)
    ollama := llm.NewOllama(cfg.OllamaHost, cfg.ModelName)
    faiss := retrieval.NewFaiss(cfg.FaissHost)

    // wire handlers
    repo := storage.NewRepository(db)
    api := &httpserver.API{
        IngestHandler:    httpserver.MakeIngestHandler(httpserver.IngestDeps{Repo: repo, LLM: ollama, EmbedModel: cfg.EmbedModel, Faiss: faiss}),
        SummarizeHandler: httpserver.MakeSummarizeHandler(httpserver.QASumDeps{Repo: repo, LLM: ollama, EmbedModel: cfg.EmbedModel, GenModel: cfg.ModelName, Faiss: faiss}),
        QAHandler:        httpserver.MakeQAHandler(httpserver.QASumDeps{Repo: repo, LLM: ollama, EmbedModel: cfg.EmbedModel, GenModel: cfg.ModelName, Faiss: faiss}),
    }
    r.Mount("/", httpserver.NewRouter(api))

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           r,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    // Graceful shutdown on SIGINT/SIGTERM
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    errCh := make(chan error, 1)
    go func() {
        log.Println("API listening on", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errCh <- err
        }
    }()

    select {
    case <-ctx.Done():
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        return srv.Shutdown(shutdownCtx)
    case err := <-errCh:
        return err
    }
}


