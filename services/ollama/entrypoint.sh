#!/bin/sh
set -e

# Run Ollama server in foreground (models đã được pull sẵn lúc build)
exec ollama serve


