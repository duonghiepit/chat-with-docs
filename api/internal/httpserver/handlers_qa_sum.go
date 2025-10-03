package httpserver

import (
    "encoding/json"
    "net/http"
    "strings"
    "context"
    "time"
    "strconv"

    "github.com/hiepdt/contest/services/api/internal/llm"
    "github.com/hiepdt/contest/services/api/internal/retrieval"
    "github.com/hiepdt/contest/services/api/internal/storage"
)

type QASumDeps struct {
    Repo *storage.Repository
    LLM  *llm.OllamaClient
    EmbedModel string
    GenModel   string
    Faiss *retrieval.FaissClient
}

func MakeSummarizeHandler(deps QASumDeps) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req SummarizeRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil { w.WriteHeader(http.StatusBadRequest); return }
        ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
        defer cancel()
        // Lấy vài chunk đầu của tài liệu để tóm tắt
        chunks, _ := deps.Repo.GetChunksByDocument(ctx, req.DocumentID, 8)
        // Đảo lại theo thời gian (mới nhất trước) -> giữ thứ tự tự nhiên
        for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 { chunks[i], chunks[j] = chunks[j], chunks[i] }
        joined := strings.Join(chunks, "\n\n")
        if len(chunks) == 0 {
            w.WriteHeader(http.StatusBadRequest)
            _ = json.NewEncoder(w).Encode(map[string]string{"error":"document_id không có dữ liệu; hãy ingest trước"})
            return
        }
        n := req.NumBullets
        if n <= 0 { n = 5 }
        cat := req.Category
        //if cat == "" { cat = "Kết luận, rủi ro" }
		userInst := strings.TrimSpace(req.Instruction)
        //if userInst == "" { userInst = "Tóm tắt kết luận và rủi ro chính" }
        prompt := "Bạn là chuyên gia kinh tế tài chính. Nhiệm vụ: " + userInst + "và tạo đúng " + strconv.Itoa(n) + " gạch đầu dòng ngắn gọn (mỗi gạch 20 đến 25 từ) có ý nghĩa và insight sâu sắc từ việc summarize tài liệu giúp người dùng có thể dễ dàng hiểu được, danh mục cần tập trung là: " + cat + ". Cuối cùng đưa ra kết luận nhé.\n" +
            "Chỉ dùng thông tin trong văn bản cung cấp.\n" +
            "Xuất duy nhất JSON theo mẫu: {\"bullets\":[\"...\"]} với đúng " + strconv.Itoa(n) + " phần tử, không thêm tiền tố hay lời dẫn.\n" +
            "Văn bản:\n" + joined
        out, err := deps.LLM.Generate(ctx, prompt)
        if err != nil { w.WriteHeader(500); return }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        // Thử parse JSON theo schema yêu cầu
        var parsed struct{ Bullets []string `json:"bullets"` }
        err = json.Unmarshal([]byte(out), &parsed)
        lines := []string{}
        if err == nil && len(parsed.Bullets) > 0 {
            lines = parsed.Bullets
        } else {
            // Fallback: chuẩn hoá theo từng dòng
            for _, ln := range strings.Split(strings.TrimSpace(out), "\n") {
                ln = strings.TrimSpace(strings.TrimLeft(ln, "-•* "))
                if ln != "" && !strings.EqualFold(ln, "gạch đầu dòng:") { lines = append(lines, ln) }
            }
        }
        if len(lines) > n { lines = lines[:n] }
        resp := map[string]any{
            "sections": []map[string]any{{"title": cat, "bullets": lines}},
            "citations": []any{},
            "meta": map[string]any{"model": deps.GenModel, "prompt_tokens": 0, "completion_tokens": 0, "latency_ms": 0},
        }
        _ = json.NewEncoder(w).Encode(resp)
    }
}

func MakeQAHandler(deps QASumDeps) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req QARequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil { w.WriteHeader(http.StatusBadRequest); return }
        if req.TopK <= 0 { req.TopK = 5 }
        ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
        defer cancel()
        embeds, err := deps.LLM.Embeddings(ctx, deps.EmbedModel, []string{req.Question})
        if err != nil || len(embeds) == 0 { w.WriteHeader(500); return }
        var hits []struct{ID int64; DocID string; Content string; Score float32}
        if deps.Faiss != nil {
            _, _, err := deps.Faiss.Search(ctx, embeds[0], req.TopK)
            if err == nil {
                // Chưa map id->content nên tạm thời dùng fallback DB ở dưới
            }
        }
        if len(hits) == 0 {
            // Giới hạn theo document hiện hành nếu client gửi kèm bằng cách đặt câu hỏi dạng: [doc:<id>] ...
            docScoped := ""
            if strings.HasPrefix(strings.TrimSpace(req.Question), "[doc:") {
                if p := strings.Index(req.Question, "]"); p > 5 {
                    docScoped = req.Question[5:p]
                }
            }
            if docScoped != "" {
                hits, err = deps.Repo.SimilarChunksByDoc(ctx, docScoped, embeds[0], req.TopK)
            } else {
                hits, err = deps.Repo.SimilarChunks(ctx, embeds[0], req.TopK)
            }
        }
        if err != nil { w.WriteHeader(500); return }
        var contextStr strings.Builder
        for _, h := range hits { contextStr.WriteString("- "); contextStr.WriteString(h.Content); contextStr.WriteString("\n") }
        prompt := "Bạn là trợ lý tài chính. Dựa trên ngữ cảnh sau, trả lời ngắn gọn, trích dẫn các đoạn liên quan cuối câu theo dạng [#id].\nNgữ cảnh:\n" + contextStr.String() + "\nCâu hỏi: " + req.Question
        ans, err := deps.LLM.Generate(ctx, prompt)
        if err != nil { w.WriteHeader(500); return }
        _ = json.NewEncoder(w).Encode(map[string]any{"answer": strings.TrimSpace(ans), "citations": hits})
    }
}


