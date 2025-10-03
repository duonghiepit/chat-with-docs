package retrieval

// NOTE: For macOS (Apple Silicon) there is no official FAISS CGO prebuilt.
// We will shell out to a small Python helper via subprocess later if needed,
// but here we keep an interface so the rest of code compiles.

import (
    "errors"
)

type Index interface {
    Add(vectors [][]float32, ids []int64) error
    Search(query []float32, topK int) ([]int64, []float32, error)
}

type NoopIndex struct{}

func NewNoopIndex() *NoopIndex { return &NoopIndex{} }

func (n *NoopIndex) Add(vectors [][]float32, ids []int64) error { return nil }

func (n *NoopIndex) Search(query []float32, topK int) ([]int64, []float32, error) {
    return nil, nil, errors.New("FAISS not wired yet")
}


