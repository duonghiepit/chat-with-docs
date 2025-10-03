FROM ollama/ollama:latest AS base
ENV OLLAMA_HOST=0.0.0.0

# Pre-pull models at build time into layer cache
RUN ollama serve & \
    for i in $(seq 1 120); do wget -qO- http://127.0.0.1:11434/api/tags >/dev/null 2>&1 && break || sleep 1; done && \
    ollama pull qwen2.5:3b && \
    ollama pull bge-m3 || true

FROM base
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
EXPOSE 11434
ENTRYPOINT ["/bin/ollama","serve"]