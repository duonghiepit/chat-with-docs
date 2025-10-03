package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type OllamaClient struct {
    host      string
    modelName string
    httpc     *http.Client
}

func NewOllama(host, model string) *OllamaClient {
    return &OllamaClient{
        host:      host,
        modelName: model,
        httpc: &http.Client{Timeout: 120 * time.Second},
    }
}

type generateRequest struct {
    Model  string            `json:"model"`
    Prompt string            `json:"prompt"`
    Stream bool              `json:"stream"`
    Options map[string]any   `json:"options,omitempty"`
}

type generateResponse struct {
    Response string `json:"response"`
    Done     bool   `json:"done"`
}

func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
    reqBody := generateRequest{Model: c.modelName, Prompt: prompt, Stream: false}
    b, _ := json.Marshal(reqBody)
    url := fmt.Sprintf("%s/api/generate", c.host)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.httpc.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    var out generateResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "", err }
    return out.Response, nil
}

type embedRequest struct {
    Model string   `json:"model"`
    Input []string `json:"input"`
}

type embedResponse struct {
    Embeddings [][]float32 `json:"embeddings"`
}

func (c *OllamaClient) Embeddings(ctx context.Context, model string, input []string) ([][]float32, error) {
    if model == "" { model = "nomic-embed-text" }
    reqBody := embedRequest{Model: model, Input: input}
    b, _ := json.Marshal(reqBody)
    url := fmt.Sprintf("%s/api/embeddings", c.host)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.httpc.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    var out embedResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
    return out.Embeddings, nil
}


