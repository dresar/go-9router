# Go 9Router ⚡

**Go 9Router** is an ultra-fast, lightweight, production-grade rewrite of [9Router](https://github.com/decolua/9router) in **Golang**. It serves as a unified AI Gateway, reverse proxy, and model aggregator for OpenAI, Anthropic, Gemini, DeepSeek, xAI Grok, Groq, and custom local models.

---

## ✨ Features

- **Blazing Fast Performance**: Zero-overhead request routing in Golang (<15ms cold start, ~12MB RAM idle).
- **Single Binary Distribution**: Compiles to a single standalone executable (`bin/9router.exe`) containing the gateway server daemon, the CLI suite, process supervisor, and API client.
- **Unified CLI**: Built-in commands for `start`, `stop`, `status`, interactive terminal menu (`menu`), and `xai video` generation.
- **Isolated Frontend**: Extracted Next.js 16 + React 19 web dashboard located in `frontend/`, with automatic proxying through the Go backend on `http://localhost:20128/dashboard`.
- **Decoupled Agent Skills**: Ready-to-use drop-in skills for Claude Code, Cursor, and custom AI agents in `skills/`.
- **Robust SQLite Storage**: Pure-Go SQLite engine (`modernc.org/sqlite` without CGO requirement) with automated WAL migrations for providers, keys, combos, proxy pools, and usage stats.
- **Provider Fallback & Account Rotation**: Automatic model lock detection, failover to backup accounts, and error backoff.

---

## 🚀 Quick Start

### 1. Build Binaries
```powershell
.\build.ps1
```
This produces `bin/9router.exe` (~11.8 MB).

### 2. Start Gateway
```powershell
.\bin\9router.exe
```
This automatically initializes the database, checks for port conflicts, starts listening on `http://0.0.0.0:20128`, and opens the dashboard in your default browser.

### 3. Check Status
```powershell
.\bin\9router.exe status
```

### 4. Interactive Terminal Menu
```powershell
.\bin\9router.exe menu
```

### 5. Generate Grok Video via CLI
```powershell
.\bin\9router.exe xai video --prompt "Cinematic drone view of a cyberpunk city in rainy night" --output city.mp4
```

---

## 🎨 Next.js Web Dashboard

The web dashboard is located in `frontend/`. To launch the Next.js dev server:
```powershell
cd frontend
bun run dev
```
It runs on port `20127`, and the Go gateway at port `20128` reverse-proxies dashboard traffic directly to it!

---

## 📁 Project Structure

```
go-9router/
├── cmd/
│   ├── 9router/              # Unified CLI and Server binary
│   └── gateway/              # Direct daemon binary
├── internal/
│   ├── auth/                 # JWT session & bcrypt password hashing
│   ├── cli/                  # CLI commands, xai video, process supervisor
│   ├── config/               # Environment & configuration loader
│   ├── handlers/             # REST API & OpenAI v1 endpoints
│   ├── logging/              # Structured high-throughput logger
│   ├── middleware/           # CORS, session auth, latency logger
│   ├── providers/            # Provider adapters & credential selection
│   ├── router/               # Unified HTTP multiplexer & frontend proxy
│   ├── server/               # HTTP server lifecycle
│   ├── storage/              # SQLite database schema & repositories
│   └── stream/               # SSE & JSON streaming proxies
├── frontend/                 # Extracted Next.js 16 Web Dashboard
├── skills/                   # Isolated AI Agent Skills (Claude, Cursor, etc.)
├── ARCHITECTURE.md           # Deep architectural analysis and audit
├── build.ps1                 # One-click build script
└── README.md
```

---

## 📄 License
MIT License
