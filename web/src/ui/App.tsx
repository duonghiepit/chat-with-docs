import React, { useEffect, useMemo, useRef, useState } from 'react'

const API = (import.meta.env.VITE_API_URL as string) || `${window.location.protocol}//${window.location.hostname}:8080`

type Citation = { source: string; page: number; span: [number, number] }

export default function App() {
  const [tab, setTab] = useState<'summary'|'qa'>('summary')
  const [numBullets, setNumBullets] = useState<number>(0)
  const [category, setCategory] = useState<string>('')
  const [file, setFile] = useState<File | null>(null)
  const [text, setText] = useState('')
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(false)
  const [summary, setSummary] = useState<any>(null)
  const [qaAnswer, setQaAnswer] = useState<any>(null)
  const [sessionId] = useState<string>(() => Math.random().toString(36).slice(2))
  const [history, setHistory] = useState<any[]>([])

  useEffect(() => {}, [])

  const fetchWithTimeout = async (input: RequestInfo, init: RequestInit = {}, ms = 20000) => {
    const controller = new AbortController()
    const id = setTimeout(() => controller.abort(), ms)
    try {
      const res = await fetch(input, { ...init, signal: controller.signal })
      return res
    } finally {
      clearTimeout(id)
    }
  }

  const [documentId, setDocumentId] = useState<string>('')

  const onIngest = async (): Promise<boolean> => {
    try {
      if (file) {
        alert('Upload PDF chưa được hỗ trợ ở backend hiện tại. Vui lòng dán văn bản.')
      }
      if (text.trim()) {
        const did = documentId || `doc-${Date.now()}`
        const chunks = text
          .split(/\n\n+/)
          .map(s => s.trim())
          .filter(Boolean)
        const payload = { document_id: did, chunks }
        const r = await fetchWithTimeout(`${API}/ingest`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        })
        if (!r.ok) throw new Error(`ingest failed: ${r.status}`)
        setDocumentId(did)
      }
      return true
    } catch (e) {
      console.error(e)
      return false
    }
  }

  const onSummarize = async () => {
    setSummary(null)
    setQaAnswer(null)
    setLoading(true)
    try {
      // Tự động ingest nếu có file/text mới
      if (file || text.trim()) {
        await onIngest()
      }
      const did = documentId || `doc-${Date.now()}`
      const res = await fetch(`${API}/summarize`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ document_id: did, num_bullets: numBullets, category, instruction: query })
      })
      const js = await res.json()
      setSummary(js)
      setHistory(h => [{ ts: Date.now()/1000, type: 'summary', query }, ...h].slice(0,50))
    } finally {
      setLoading(false)
    }
  }

  const onQA = async () => {
    setQaAnswer(null)
    setSummary(null)
    setLoading(true)
    try {
      // Tự động ingest nếu có file/text mới
      if (file || text.trim()) {
        await onIngest()
      }
      const res = await fetch(`${API}/qa`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ question: query, top_k: 5 })
      })
      const js = await res.json()
      setQaAnswer(js)
      setHistory(h => [{ ts: Date.now()/1000, type: 'qa', query }, ...h].slice(0,50))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ maxWidth: 1200, margin: '16px auto', padding: '0 12px', fontFamily: 'Inter, system-ui, sans-serif' }}>
      <h2 className="title">Trợ lý AI truy vấn nội dung báo cáo kinh tế tài chính</h2>
      <div className="layout">
        <div className="stack">
          <div className="card">
          <div className="tabs">
            <div className={`tab ${tab==='summary'?'active':''}`} onClick={() => setTab('summary')}>Tóm tắt</div>
            <div className={`tab ${tab==='qa'?'active':''}`} onClick={() => setTab('qa')}>Hỏi-Đáp (QA)</div>
          </div>
          <div style={{ marginBottom: 10 }}>
            <label>Tải lên (PDF) hoặc dán văn bản</label>
            <div className='file'>
              <label className='pick'>Chọn tệp PDF
                <input type='file' accept='.pdf' onChange={e => setFile(e.target.files?.[0] || null)} />
              </label>
              <div className='name'>{file ? file.name : 'Chưa chọn tệp'}</div>
            </div>
            <textarea placeholder='Hoặc dán văn bản tóm tắt vào đây...' value={text} onChange={e => setText(e.target.value)} />
          </div>
          <div className="controls">
            <div>
              <label>Số lượng gạch đầu dòng</label>
              <input type='number' value={numBullets || ''} placeholder='Ví dụ: 5' onChange={e => setNumBullets(parseInt(e.target.value || '0'))} />
            </div>
            <div>
              <label>Danh mục cần tập trung</label>
              <input type='text' value={category} placeholder='Ví dụ: kết luận, xu hướng' onChange={e => setCategory(e.target.value)} />
            </div>
          </div>
          <div style={{ marginTop: 12 }}>
            <label>{tab==='qa'?'Câu hỏi':'Yêu cầu tóm tắt'}</label>
            <input type='text' value={query} placeholder={tab==='qa'? 'Nhập câu hỏi...' : 'Nhập yêu cầu tóm tắt...'} onChange={e => setQuery(e.target.value)} />
          </div>
          <div style={{ display: 'flex', gap: 12, marginTop: 16, flexWrap: 'wrap' }}>
            {tab==='summary' ? (
              <button onClick={onSummarize} disabled={loading}>Tạo tóm tắt</button>
            ) : (
              <button onClick={onQA} disabled={loading}>Hỏi đáp</button>
            )}
          </div>
          </div>
          <div className="card">
          <h3>Tóm tắt tài liệu</h3>
          {tab==='summary' && summary && (
            <div>
              {Array.isArray(summary.sections) && summary.sections.map((sec: any, idx: number) => (
                <div key={idx} style={{ marginBottom: 12 }}>
                  <div style={{ fontWeight: 700, marginBottom: 6 }}>{sec.title}</div>
                  <ul style={{ paddingLeft: 18, margin: 0 }}>
                    {(() => {
                      const raw: string[] = Array.isArray(sec.bullets)
                        ? (sec.bullets as string[])
                        : String(sec.bullets || '')
                            .split('\n')
                            .map((s: string) => s.trim())
                            .filter(Boolean);
                      const bullets = raw
                        .map((b: string) => b.replace(/^[-–•*]\s?/, '').trim())
                        .filter((b: string) => {
                          const lower = b.toLowerCase();
                          const isHeader = /gạch\s*đầu\s*dòng/.test(lower) || /:\s*$/.test(b) && b.split(/\s+/).length <= 4;
                          return !isHeader;
                        });
                      return bullets.map((b: string, i: number) => (
                        <li key={i} style={{ lineHeight: 1.5 }}>{b}</li>
                      ));
                    })()}
                  </ul>
                </div>
              ))}
              <div className="meta">Model: {summary.meta?.model} | Tokens: {summary.meta?.prompt_tokens}+{summary.meta?.completion_tokens} | Độ trễ: {summary.meta?.latency_ms}ms</div>
            </div>
          )}
          {tab==='qa' && qaAnswer && (
            <div>
              <div><strong>Trả lời:</strong> {qaAnswer.answer}</div>
              <div><strong>Độ tin cậy:</strong> {qaAnswer.confidence}</div>
              <h4>Nguồn trích dẫn</h4>
              <pre style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(qaAnswer.citations, null, 2)}</pre>
            </div>
          )}
          </div>
        </div>
        <div className="card">
          <h3>Lịch sử</h3>
          <div style={{ display:'flex', gap:8, marginBottom:8 }}>
            <button onClick={()=>setHistory(h=>[...h])}>Làm mới</button>
            <button onClick={()=>setHistory([])}>Xoá</button>
          </div>
          <ul style={{ maxHeight: 260, overflow:'auto' }}>
            {history.map((h:any, i:number)=> (
              <li key={i} style={{ marginBottom:8 }}>
                <div style={{ fontSize:12, color:'#6b7280' }}>{new Date((h.ts||0)*1000).toLocaleString()} • {h.type}</div>
                <div style={{ fontWeight:600 }}>{h.query}</div>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  )
}


