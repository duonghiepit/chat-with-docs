FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git && go mod download && go mod tidy && CGO_ENABLED=0 go build -o api ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/api /usr/local/bin/api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]


