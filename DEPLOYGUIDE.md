# Deployment Guide

This guide covers every way to run the Hello World Go server.

---

## Option 1 — Docker Compose (recommended)

The fastest path from zero to running.

### Prerequisites

- Docker 24+ with Docker Compose v2

### Steps

```bash
# 1. Clone
git clone <your-repo-url> hello-world
cd hello-world

# 2. Build and start (detached)
docker compose up --build -d

# 3. Verify
curl http://localhost:8080
# → Hello World
```

### What the build does

| Stage | Base image | Action |
|---|---|---|
| Builder | `golang:1.23-alpine` | Compiles `main.go`, runs UPX compression |
| Runtime | `scratch` | Copies the compressed binary only |

The final image contains **only** the binary — no shell, no libc, no OS layer.

### Verify routing

```bash
curl http://localhost:8080           # 200 Hello World
curl http://localhost:8080/other     # 404 404 page not found
curl -X POST http://localhost:8080   # 405 Method Not Allowed
```

### Logs

```bash
docker compose logs -f
```

Each request prints one line:

```
GET / 200
GET /missing 404
POST / 405
```

### Stop

```bash
docker compose down                      # stop, keep image
docker compose down --rmi all --volumes  # stop and remove everything
```

---

## Option 2 — Standalone Docker (no Compose)

```bash
# Build
docker build -t hello-app .

# Run
docker run -d -p 8080:8080 --name hello hello-app

# Test
curl http://localhost:8080

# Stop
docker stop hello && docker rm hello
```

Override the port via environment variable:

```bash
docker run -d -p 9000:9000 -e PORT=9000 --name hello hello-app
curl http://localhost:9000
```

---

## Option 3 — Local Go build (no Docker)

Useful for fast iteration during development.

### Windows

```powershell
go mod init hello   # skip if go.mod already exists
$Env:CGO_ENABLED=0; go build -ldflags="-s -w" -trimpath -o hello-world.exe .
.\hello-world.exe
```

```powershell
# In another terminal
curl http://localhost:8080
```

Custom port:

```powershell
$env:PORT="9090"; .\hello-world.exe
curl http://localhost:9090
```

### Linux / macOS

```bash
go mod init hello
go build -ldflags="-s -w" -trimpath -o hello-world .
./hello-world
```

Custom port:

```bash
PORT=9090 ./hello-world
curl http://localhost:9090
```

---

## Minimising binary size (local build)

| Method | Approx. size |
|---|---|
| `go build` (default) | ~7.5 MB |
| `+ CGO_ENABLED=0 -ldflags="-s -w" -trimpath` | **5.7 MB** |
| `+ upx --best --lzma` *(Linux/Docker only)* | ~1.5 MB |

> **Windows + UPX:** UPX is incompatible with Go 1.18+ on Windows.
> Compressed binaries exit with `0xfffffffe` at runtime due to conflicts between
> UPX's self-decompressor stub and Go's runtime PE initialisation.
> The `-ldflags="-s -w" -trimpath` build is the smallest safe Windows binary.

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | TCP port the server binds to |

---

## Health check

The server has no dedicated health endpoint, but `GET /` serves as one:

```bash
curl -sf http://localhost:8080 | grep -q "Hello World" && echo OK
```

For Docker, add this to `docker-compose.yml` if needed:

```yaml
healthcheck:
  test: ["CMD-SHELL", "wget -qO- http://localhost:8080 || exit 1"]
  interval: 10s
  timeout: 5s
  retries: 3
```

---

## Troubleshooting

### Port already in use

```bash
# Find what's using port 8080
lsof -i :8080          # Linux/macOS
netstat -ano | findstr 8080  # Windows
```

Change the host port in `docker-compose.yml`:

```yaml
ports:
  - "9090:8080"   # host:container
```

### Image not rebuilding after code change

```bash
docker compose up --build -d   # always force a rebuild
```

### Container exits immediately

```bash
docker compose logs app   # read the error output
```
