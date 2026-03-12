# Hello World — Minimal Go on Docker

A **production-style** Hello World in Go, optimised for the smallest possible
binary and image size. The Go binary serves HTTP directly — no reverse proxy.

---

## Project structure

```
hello-world/
├── main.go            # Go HTTP server (~25 lines)
├── Dockerfile         # 2-stage build → scratch + UPX (~1.5 MB image)
├── docker-compose.yml # Single service, port 8080
└── README.md
```

---

## Why so small?

| Technique | Effect |
|---|---|
| `CGO_ENABLED=0` | Fully static binary, no C runtime needed |
| `-ldflags="-s -w"` | Strips debug symbols and DWARF info |
| `-trimpath` | Removes local build paths from binary |
| `FROM scratch` | Zero-OS final image — only the binary |
| `upx --best --lzma` | LZMA compression (~60% size reduction, Linux/Docker only — incompatible with Go 1.18+ on Windows) |

Typical final image size: **~1.5 MB**.

---

## Prerequisites

| Tool | Minimum version |
|---|---|
| Docker | 24+ |
| Docker Compose | v2 (included with Docker Desktop) |

> Go is **not required** on the host — the build happens entirely inside Docker.

---

## Step-by-step deployment

### 1 — Clone the project

```bash
git clone <your-repo-url> hello-world
cd hello-world
```

### 2 — Build and start

```bash
docker compose up --build -d
```

What happens:
1. **Stage 1** — `golang:1.23-alpine` compiles `main.go` into a static binary.
2. **Stage 1** — UPX compresses the binary with `--best --lzma`.
3. **Stage 2** — The compressed binary is copied into an empty `scratch` image.

### 3 — Test it

```bash
curl http://localhost:8080
```

Expected output:

```
Hello World
```

```bash
curl http://localhost:8080/other   # → 404
curl -X POST http://localhost:8080 # → 405
```

### 4 — Inspect image sizes

```bash
docker images | grep hello
```

Example output:

```
hello-world-app   latest   1.5MB
```

### 5 — View running containers

```bash
docker compose ps
```

```
NAME        IMAGE              STATUS    PORTS
hello-app   hello-world-app    Up        0.0.0.0:8080->8080/tcp
```

### 6 — View logs

```bash
docker compose logs -f
```

Each request is logged in this format:

```
GET / 200
GET /other 404
POST / 405
```

### 7 — Stop and clean up

```bash
# Stop containers (keep images)
docker compose down

# Stop AND remove images + volumes
docker compose down --rmi all --volumes
```

---

## Running locally (without Docker)

The `Dockerfile` generates `go.mod` internally. For a local build, initialise
the module first.

### Windows

```powershell
go mod init hello
go build -ldflags="-s -w" -trimpath -o hello-world.exe .
.\hello-world.exe
```

Test in another terminal:

```powershell
curl http://localhost:8080
# → Hello World
```

### Linux / macOS

```bash
go mod init hello
go build -ldflags="-s -w" -trimpath -o hello-world .
./hello-world
```

---

## Docker image size comparison

| Build flags | Size |
|---|---|
| Default `go build` | ~7.5 MB |
| `+ CGO_ENABLED=0` | ~6.8 MB |
| `+ -ldflags="-s -w" -trimpath` | ~4.7 MB |
| `+ UPX --best --lzma` *(Linux/Docker only)* | ~1.5 MB |

---

## Windows binary size

| Step | Size |
|---|---|
| Plain `go build` | ~7.5 MB |
| `CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath` | **5.7 MB** |

> UPX is incompatible with Go 1.18+ on Windows — compressed binaries crash
> at runtime with exit code `0xfffffffe` due to conflicts between UPX's
> self-decompressor stub and Go's runtime PE initialisation.
> The strip-flags build is the smallest safe Windows binary.

---

## Architecture

```
Browser / curl
      │
      ▼  :8080
  ┌────────┐
  │  Go    │  (scratch image — static binary, ~1.5 MB)
  └────────┘
```

The Go binary serves HTTP directly on port 8080.

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port the server listens on |

---

## License

MIT
