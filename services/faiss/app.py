from fastapi import FastAPI
from pydantic import BaseModel
import faiss
import numpy as np

app = FastAPI()

# Cosine similarity via inner product on normalized vectors
dim = 768
index = faiss.IndexFlatIP(dim)

class AddItem(BaseModel):
    id: int
    vector: list[float]

class AddRequest(BaseModel):
    items: list[AddItem]

class SearchRequest(BaseModel):
    vector: list[float]
    top_k: int = 5

@app.post("/add")
def add_vectors(req: AddRequest):
    if not req.items:
        return {"added": 0}
    vecs = []
    ids = []
    for it in req.items:
        v = np.array(it.vector, dtype="float32")
        if v.shape[0] != dim:
            raise ValueError("vector dim mismatch")
        # normalize for cosine
        n = np.linalg.norm(v)
        if n > 0:
            v = v / n
        vecs.append(v)
        ids.append(it.id)
    xb = np.stack(vecs, axis=0)
    index.add_with_ids(xb, np.array(ids, dtype="int64"))
    return {"added": len(ids)}

@app.post("/search")
def search(req: SearchRequest):
    v = np.array(req.vector, dtype="float32")
    n = np.linalg.norm(v)
    if n>0:
        v = v / n
    v = v.reshape(1,-1)
    scores, ids = index.search(v, req.top_k)
    res = []
    for i in range(ids.shape[1]):
        if ids[0, i] == -1:
            continue
        res.append({"id": int(ids[0,i]), "score": float(scores[0,i])})
    return {"results": res}


