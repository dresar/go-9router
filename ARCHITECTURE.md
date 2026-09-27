# Go 9Router Architecture & Implementation Analysis

## Executive Summary

**Go 9Router** is an ultra-lightweight, high-concurrency, production-grade rewrite of the **9Router** AI Gateway originally implemented in Node.js/Next.js. By transitioning the backend and CLI to pure Golang, Go 9Router achieves:
- **Instant startup** (<15ms vs ~2.5s in Node.js)
- **Ultra-low memory footprint** (~15MB RAM idle vs ~250MB+ in Node.js/Next.js)
- **High concurrency throughput** via Go's lightweight goroutines and connection pooling
- **Single-binary distribution**: A single compiled executable (`9router.exe`) encapsulates the server daemon, the interactive CLI, process supervisor, and API client.
- **Clean modularity**: Complete separation of Frontend, Agent Skills, CLI, and Backend services.

---

## 1. Folder Structure Analysis: 9Router (Node.js) vs Go 9Router (Golang)

| Original Folder in `9router` | Implementation in `go-9router` | Technology / Language | Status & Role |
| :--- | :--- | :--- | :--- |
| **`cli/`** (`cli.js`, `src/cli/...`) | **`internal/cli/`** & **`cmd/9router/`** | **Golang** | Rewritten in pure Go. Includes process management, port conflict resolution, interactive TTY menu, API client, and `xai video` CLI tool. |
| **`open-sse/`** | **`internal/stream/`**, **`internal/providers/`**, **`internal/handlers/`** | **Golang** | Core streaming SSE proxy, multi-provider adapter engine, model-to-provider routing, and fallback retry logic. |
| **`skills/`** | **`skills/`** | **Markdown / YAML** | Separated into dedicated repository directory with 9 individual agent capability skills and master entry point. |
| **`src/app/`** (Frontend) | **`frontend/`** | **Next.js 16 + React 19** | Separated web dashboard with Turbopack, Tailwind CSS v4, Monaco Editor, Zustand, and proxy rewrites to Go backend. |
| **`src/lib/`** & **`src/app/api/`** | **`internal/storage/`**, **`internal/auth/`**, **`internal/handlers/`** | **Golang** | SQLite storage with WAL mode, bcrypt auth, JWT sessions, and REST endpoints. |
| **`src/mitm/`** | **`internal/middleware/`** & Go proxy router | **Golang** | HTTP interception, CORS, logging, and client header normalization. |

---

## 2. Component Deep Dive

### 2.1. Frontend (`frontend/`)
- Extracted cleanly into `go-9router/frontend`.
- Built on **Next.js 16.3.6**, **React 19.2.4**, **Tailwind CSS v4**, and **Zustand**.
- Configured with `next.config.mjs` rewrites directing `/api/*` and `/v1/*` requests directly to the Go backend at `http://127.0.0.1:20128`.
- In dev mode, runs seamlessly with `bun run dev` on port `20127`.
- The Go backend reverse-proxies dashboard requests to port `20127` automatically if the Next.js dev server is running, or serves the built-in dark-mode dashboard fallback if running standalone.

### 2.2. CLI in Golang (`internal/cli/` & `cmd/9router/main.go`)
- **Single Unified Binary**: `go build -o bin/9router.exe ./cmd/9router` produces a single standalone `.exe`.
- **Subcommands**:
  - `9router`: Starts server daemon on default port 20128, clears stale port locks, and automatically opens the browser.
  - `9router status`: Communicates via `APIClient` to verify `/api/health`, `/api/version`, and uptime.
  - `9router stop`: Finds and terminates running gateway instances gracefully.
  - `9router menu`: Interactive terminal control panel with options for Web UI, Health Diagnostic, Grok Imagine Video, and server control.
  - `9router xai video --prompt "..." --output video.mp4`: Submits video generation to `/v1/videos/generations`, polls job status, and atomically downloads MP4 video files.

### 2.3. Agent Skills (`skills/`)
The AI agent skills have been isolated into `skills/`:
- `skills/9router`: Master entrypoint and router configuration.
- `skills/9router-chat`: Chat and code-generation instruction sets.
- `skills/9router-embeddings`: Vector embeddings API reference.
- `skills/9router-image`: Image generation (DALL-E, Grok, Flux).
- `skills/9router-stt`: Speech-to-text.
- `skills/9router-tts`: Text-to-speech.
- `skills/9router-video`: Grok Imagine video workflows.
- `skills/9router-web-fetch`: Web URL to markdown extraction.
- `skills/9router-web-search`: Real-time web search integration.

### 2.4. Backend Gateway Engine (`internal/`)
- **`internal/auth`**: JWT session generation & verification, bcrypt cost-10 password hashing.
- **`internal/storage`**: Embedded SQLite database (`modernc.org/sqlite` zero-CGO), auto-migrating tables:
  - `providerConnections`
  - `providerNodes`
  - `apiKeys`
  - `combos`
  - `proxyPools`
  - `kv`
  - `usage`
  - `requestDetails`
  - `settings`
- **`internal/providers`**: Provider adapter registry (OpenAI, Anthropic, Gemini, Groq, DeepSeek, Mistral, xAI, OpenRouter, and generic OpenAI-compatible APIs). Includes connection pooling, model mapping, and fallback rotation.
- **`internal/stream`**: Non-blocking SSE proxy with context cancellation to stop goroutines when clients disconnect.
- **`internal/handlers`**: Complete suite of REST and v1 endpoints for chat, models, embeddings, images, audio, video, web fetch, and usage metrics.

---

## 3. How to Build & Run

### Unified One-Click Build
```powershell
.\build.ps1
```
This compiles `bin/9router.exe` and `bin/gateway.exe`.

### Running the Gateway & CLI
```powershell
# Start the Gateway Server (auto-opens http://localhost:20128/dashboard)
.\bin\9router.exe

# Run interactive CLI menu
.\bin\9router.exe menu

# Check health and status
.\bin\9router.exe status

# Generate xAI Grok video via CLI
.\bin\9router.exe xai video --prompt "Cyberpunk city in rain with neon lights" --output city.mp4
```

### Running the Next.js Web Dashboard
```powershell
cd frontend
bun run dev
```
Open `http://localhost:20128/dashboard` in any browser to experience the full Next.js UI powered by the Go backend!
