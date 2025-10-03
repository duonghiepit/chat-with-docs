# Tóm tắt & QA báo cáo tài chính (Go + FAISS + Ollama)

## Yêu cầu
- Docker + Docker Compose

## Dịch vụ
- API (Go): `services/api`
- FAISS (Python/FastAPI): `services/faiss`
- Postgres 16 (+ pgvector), Redis 7, Ollama (CPU)

## Chạy nhanh
```bash
# 1) Khởi động toàn bộ stack
docker compose up -d --build

# 2) Tải model trong container Ollama (lần đầu)
docker exec -it ollama ollama pull llama3.1:8b
docker exec -it ollama ollama pull nomic-embed-text

# 3) Kiểm tra health
curl -s http://localhost:8080/health
```

## Biến môi trường chính (đặt sẵn trong docker-compose)
- `POSTGRES_URL`, `REDIS_ADDR`
- `OLLAMA_HOST`, `MODEL_NAME` (mặc định `llama3.1:8b`)
- `EMBED_MODEL` (mặc định `nomic-embed-text`)
- `FAISS_HOST` (mặc định `http://faiss:8000` trong compose)

## API
### 1) Ingest tài liệu
- Tạo embeddings qua Ollama, lưu Postgres, thêm vector vào FAISS
```bash
curl -X POST http://localhost:8080/ingest \
  -H 'Content-Type: application/json' \
  -d '{
    "document_id": "doc-001",
    "chunks": [
      "Doanh thu quý 2 tăng 12% YoY...",
      "Biên lợi nhuận gộp cải thiện 1.8 điểm % ..."
    ]
  }'
```

### 2) Hỏi đáp (RAG)
- Embed câu hỏi, search FAISS top-k, gọi LLM sinh câu trả lời
```bash
curl -X POST http://localhost:8080/qa \
  -H 'Content-Type: application/json' \
  -d '{
    "question": "Tăng trưởng doanh thu quý 2 là bao nhiêu?",
    "top_k": 5
  }'
```

### 3) Tóm tắt
```bash
curl -X POST http://localhost:8080/summarize \
  -H 'Content-Type: application/json' \
  -d '{"document_id": "doc-001"}'
```

### 4) Metrics
```bash
curl -s http://localhost:8080/metrics
```

## Ghi chú triển khai
- FAISS chạy cosine (chuẩn hoá vector trước khi add/search).
- Nếu FAISS lỗi, backend fallback truy vấn tương tự bằng `pgvector`.
- Bảng: `documents`, `chunks(embedding VECTOR(768))`, `audits`.

## Phát triển
```bash
# Chạy riêng từng service nếu cần
( cd services/faiss && docker build -t local/faiss . )
( cd services/api && docker build -t local/api . )
```


