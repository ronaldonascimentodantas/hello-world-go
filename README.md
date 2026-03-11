# Hello World — Minimal Go + Nginx on Docker

A **production-style** Hello World in Go, optimised for the smallest possible
binary and image size, served behind Nginx as a reverse proxy.

---

## Project structure

```
hello-world/
├── main.go            # Go source (7 lines)
├── Dockerfile         # Multi-stage build → scratch image
├── nginx.conf         # Nginx reverse-proxy config
├── docker-compose.yml # Orchestrates app + nginx
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

Typical final image size: **~6 MB** (Go binary only, no OS at all).

---

## Prerequisites

| Tool | Minimum version |
|---|---|
| Docker | 24+ |
| Docker Compose | v2 (included with Docker Desktop) |

> Go is **not required** on the host — the build happens entirely inside Docker.

---

## Step-by-step deployment

### 1 — Clone / copy the project

```bash
git clone <your-repo-url> hello-world
cd hello-world
```

Or just copy the four files into an empty directory.

---

### 2 — Build and start

```bash
docker compose up --build -d
```

What happens:
1. **Stage 1** — `golang:1.23-alpine` compiles `main.go` into a static binary.
2. **Stage 2** — The binary is copied into a `scratch` (empty) image.
3. `nginx:alpine` starts and proxies port **80 → app:8080**.

---

### 3 — Test it

```bash
curl http://localhost
```

Expected output:

```
Hello World
```

Or open `http://localhost` in your browser.

---

### 4 — Inspect image sizes

```bash
docker images | grep hello
```

Example output:

```
hello-world-app   latest   6.2MB
nginx             alpine   43MB
```

---

### 5 — View running containers

```bash
docker compose ps
```

```
NAME           IMAGE              STATUS    PORTS
hello-app      hello-world-app    Up        8080/tcp
hello-nginx    nginx:alpine       Up        0.0.0.0:80->80/tcp
```

---

### 6 — View logs

```bash
# All services
docker compose logs -f

# Only the Go app
docker compose logs -f app

# Only nginx
docker compose logs -f nginx
```

---

### 7 — Stop and clean up

```bash
# Stop containers (keep images)
docker compose down

# Stop AND remove images + volumes
docker compose down --rmi all --volumes
```

---

## Running locally (without Docker)

The `Dockerfile` generates `go.mod` internally during the container build, so
it is not committed to the repo. For a local build you must initialise the
module first.

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

> **Note:** if you want to commit `go.mod` so future contributors skip this
> step, that is perfectly fine — it is a one-liner file with no downsides.

---

## Minimising the Windows `.exe` size

| Method | Size |
|---|---|
| `go build` (default) | ~7.5 MB |
| `+ -ldflags="-s -w" -trimpath` | ~4.7 MB |
| `+ upx --best` *(Windows-safe)* | ~2.2 MB |
| `+ upx --best --lzma` *(Linux only)* | ~1.5 MB |

### Step 1 — Strip debug symbols (~4.7 MB)

```powershell
go build -ldflags="-s -w" -trimpath -o hello-world.exe .
```

### Step 2 — Compress with UPX (~2.2 MB, optional)

Install UPX:

```powershell
winget install upx
# or download from https://github.com/upx/upx/releases
```

Then compress:

```powershell
upx --best hello-world.exe
```

> ⚠️ **Windows caveat:** always use `--best` on Windows. The `--lzma` flag
> works inside Linux containers but corrupts the Windows PE executable loader
> and causes the binary to fail with error `0xfffffffe`. Never use
> `--best --lzma` for Windows targets.

---

## Running without Compose (standalone Docker)

Build the image manually:

```bash
docker build -t hello-app .
```

Run the Go app directly (no nginx):

```bash
docker run -d -p 8080:8080 --name hello hello-app
curl http://localhost:8080
```

---

## Architecture

```
Browser / curl
      │
      ▼  :80
  ┌────────┐
  │  Nginx │  (nginx:alpine — reverse proxy)
  └────────┘
      │
      ▼  :8080  (internal Docker network)
  ┌────────┐
  │  Go    │  (scratch image — static binary)
  └────────┘
```

Nginx is the only container exposing a port to the host.  
The Go container is internal, unreachable from outside.

---

## Customisation

### Change the port

Edit `docker-compose.yml`:

```yaml
ports:
  - "3000:80"   # expose on host port 3000 instead
```

### Change the response message

Edit `main.go`, then rebuild:

```bash
docker compose up --build -d
```

### Add HTTPS (TLS)

Replace `nginx.conf` with a config that includes:

```nginx
listen 443 ssl;
ssl_certificate     /etc/nginx/certs/fullchain.pem;
ssl_certificate_key /etc/nginx/certs/privkey.pem;
```

And mount your certificates in `docker-compose.yml`.

---

## Docker image size comparison

| Build flags | Size |
|---|---|
| Default `go build` | ~7.5 MB |
| `CGO_ENABLED=0` | ~6.8 MB |
| + `-ldflags="-s -w"` | ~4.8 MB |
| + `-trimpath` | ~4.7 MB |
| + UPX `--best --lzma` *(Linux/Docker only)* | ~1.5 MB |

> To enable UPX inside the Docker build (Linux — safe to use `--lzma` here),
> add this to the builder stage in `Dockerfile`:
> ```dockerfile
> RUN apk add --no-cache upx && upx --best --lzma server
> ```

---

## License

MIT
