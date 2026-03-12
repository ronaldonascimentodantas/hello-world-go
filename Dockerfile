# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY main.go .

RUN go mod init hello && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o server . && \
    apk add --no-cache upx && \
    upx --best --lzma server

# ── Stage 2: Runtime (scratch = zero OS layer) ────────────────────────────────
FROM scratch

COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
