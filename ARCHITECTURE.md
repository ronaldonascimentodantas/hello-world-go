# Architecture

## Overview

A single Go binary serves HTTP directly on port 8080. There is no reverse
proxy, no sidecar, and no OS layer in the container image.

```
Browser / curl
      │
      ▼  :8080
  ┌──────────────────────────────────┐
  │  Go HTTP server                  │
  │  (scratch image — ~1.5 MB)       │
  │                                  │
  │  GET /        → 200 Hello World  │
  │  GET /other   → 404              │
  │  POST /       → 405              │
  └──────────────────────────────────┘
```

---

## Request routing

Go's default `http.ServeMux` uses prefix matching. Registering `"/"` catches
every path, so routing is handled manually inside the handler:

```
Incoming request
       │
       ├─ r.URL.Path != "/"  →  404 Not Found
       │
       ├─ r.Method != GET    →  405 Method Not Allowed
       │
       └─ GET /              →  200 Hello World
```

Logging happens before each response:

```
METHOD /path STATUS
```

Example:

```
GET / 200
GET /favicon.ico 404
POST / 405
```

---

## Docker image layers

The build uses a 2-stage Dockerfile.

### Stage 1 — Builder (`golang:1.23-alpine`)

| Step | Purpose |
|---|---|
| `go mod init hello` | Create a minimal module (no `go.mod` committed) |
| `CGO_ENABLED=0` | Disable cgo → fully static binary, no libc dependency |
| `GOOS=linux GOARCH=amd64` | Target Linux/amd64 regardless of the host OS |
| `-ldflags="-s -w"` | Strip debug symbols and DWARF tables |
| `-trimpath` | Remove host filesystem paths from the binary |
| `apk add upx` + `upx --best --lzma` | LZMA-compress the binary (~60% reduction) |

### Stage 2 — Runtime (`scratch`)

`scratch` is a reserved Docker name for an empty image — no shell, no
filesystem, no OS. The compressed binary is the only file in the image.

```
Final image contents:
  /server   ← the Go binary (compressed with UPX)
```

Image size: **~1.5 MB**

---

## Binary size breakdown

| Flags applied cumulatively | Size |
|---|---|
| `go build` (default) | ~7.5 MB |
| `+ CGO_ENABLED=0` | ~6.8 MB |
| `+ -ldflags="-s -w"` | ~4.8 MB |
| `+ -trimpath` | ~4.7 MB |
| `+ upx --best --lzma` | ~1.5 MB |

---

## Port configuration

The server reads the `PORT` environment variable at startup. If unset, it
falls back to `8080`. This follows the [12-factor app](https://12factor.net/port-binding)
convention and makes the container easy to run on platforms that assign ports
dynamically (Cloud Run, Fly.io, Railway, etc.).

```
PORT env var set?
       │
       ├─ yes  →  listen on $PORT
       └─ no   →  listen on 8080
```

---

## What was removed (vs prior Nginx setup)

| Component | Before | After |
|---|---|---|
| Reverse proxy | `nginx:alpine` (43 MB) | none |
| Port exposed to host | 80 (nginx) | 8080 (Go direct) |
| Total image count | 2 (app + nginx) | 1 (app only) |
| Total image size | ~50 MB | ~1.5 MB |
| `nginx.conf` | required | deleted |
| Docker network | internal bridge (app ↔ nginx) | not needed |

---

## Sequence diagram

```
Client          Docker host         Go binary
  │                  │                  │
  │── GET / ────────►│                  │
  │                  │── forward ──────►│
  │                  │                  │── log: GET / 200
  │                  │◄── 200 ──────────│
  │◄── 200 ──────────│                  │
  │                  │                  │
  │── GET /x ───────►│                  │
  │                  │── forward ──────►│
  │                  │                  │── log: GET /x 404
  │                  │◄── 404 ──────────│
  │◄── 404 ──────────│                  │
```
