package retrieval

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type FaissClient struct {
    host string
    httpc *http.Client
}

func NewFaiss(host string) *FaissClient {
    return &FaissClient{host: host, httpc: &http.Client{Timeout: 30 * time.Second}}
}

type addItem struct { ID int64 `json:"id"`; Vector []float32 `json:"vector"` }
type addReq struct { Items []addItem `json:"items"` }
type addRes struct { Added int `json:"added"` }

func (c *FaissClient) Add(ctx context.Context, items map[int64][]float32) error {
    arr := make([]addItem, 0, len(items))
    for id, v := range items { arr = append(arr, addItem{ID: id, Vector: v}) }
    b, _ := json.Marshal(addReq{Items: arr})
    url := fmt.Sprintf("%s/add", c.host)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.httpc.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 { return fmt.Errorf("faiss add status %d", resp.StatusCode) }
    return nil
}

type searchReq struct { Vector []float32 `json:"vector"`; TopK int `json:"top_k"` }
type searchRes struct { Results []struct{ ID int64 `json:"id"`; Score float32 `json:"score"` } `json:"results"` }

func (c *FaissClient) Search(ctx context.Context, vector []float32, topK int) ([]int64, []float32, error) {
    b, _ := json.Marshal(searchReq{Vector: vector, TopK: topK})
    url := fmt.Sprintf("%s/search", c.host)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.httpc.Do(req)
    if err != nil { return nil, nil, err }
    defer resp.Body.Close()
    var out searchRes
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, nil, err }
    ids := make([]int64, 0, len(out.Results))
    scores := make([]float32, 0, len(out.Results))
    for _, r := range out.Results { ids = append(ids, r.ID); scores = append(scores, r.Score) }
    return ids, scores, nil
}


