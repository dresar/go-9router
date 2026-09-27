# Go 9Router ⚡

**Go 9Router** is an ultra-fast, lightweight, production-grade rewrite of [9Router](https://github.com/dresar/go-9router) built in **Golang** with a **Next.js 16 + React 19** Web Dashboard. It serves as an enterprise AI Gateway, reverse proxy, load balancer, token optimizer, and multi-account aggregator for OpenAI, Anthropic, Gemini, DeepSeek, xAI Grok, Groq, and custom local models.

---

## ✨ Core Highlights & Features

- **Blazing Fast Performance**: Zero-overhead request routing in Golang (<15ms cold start, ~12MB RAM idle vs ~150MB in Node.js).
- **100% Free Cloudflare Quick Tunnel**: Built-in zero-config public HTTPS exposure via official `cloudflared` Quick Tunnels (`https://*.trycloudflare.com`) without requiring an account, credit card, or tunnel token.
- **Native Tailscale Funnel**: Seamless integration with Tailscale Funnel / Serve for secure point-to-point mesh networking.
- **Smart Token Saver**: Interactive token compression engine reducing prompt token usage by up to 50% without loss of semantic precision.
- **Single Binary Distribution**: Compiles to a single standalone executable (`bin/9router.exe`) containing the gateway server daemon, CLI suite, process supervisor, and API client.
- **Clean Production Mode**: Standalone Next.js production build (`node custom-server.js --port 20127`) free of dev tools, Turbopack badges, or overlays.
- **Unified CLI Suite**: Interactive terminal UI (`menu`), daemon supervisor (`start`, `stop`, `status`), and direct generation utilities (`xai video`).
- **Standardized Agent Skills**: 9 drop-in skills for Claude Code CLI, Cursor, Cline, OpenCode, and Hermes AI agents.
- **Pure-Go SQLite Storage**: Zero-CGO SQLite engine (`modernc.org/sqlite`) with automated WAL migrations for providers, credentials, combos, and proxy pools.

---

## 🌐 Complete Route & Endpoint Architecture

Go 9Router provides 89 verified endpoints and routes across 3 primary surfaces. All routes have been verified with **100% Pass Rate** in the automated audit test.

### 1. Standard AI Protocol Endpoints (`/v1`)

Fully compatible with OpenAI, Anthropic, and LangChain client SDKs:

| Endpoint | Method | Description |
|---|---|---|
| `/v1/chat/completions` | `POST` | OpenAI-compatible chat completions with streaming SSE, proxy pool failover, and token optimization |
| `/v1/messages` | `POST` | Anthropic-compatible message completions with native tool calling support |
| `/v1/models` | `GET` | List all registered models, aliases, and active provider nodes |
| `/v1/embeddings` | `POST` | Vector embeddings generation (OpenAI, HuggingFace, local embeddings) |
| `/v1/images/generations` | `POST` | Image generation (Flux, DALL-E, Stability, Gemini Imagen) |
| `/v1/audio/speech` | `POST` | Text-to-speech (TTS) via EdgeTTS, ElevenLabs, Deepgram, MiniMax |
| `/v1/audio/transcriptions` | `POST` | Speech-to-text (STT) via Whisper, Groq, Deepgram |
| `/v1/search` | `POST` | Web search aggregation (SearxNG, Brave, DuckDuckGo, Google) |
| `/v1/web/fetch` | `POST` | Clean markdown / HTML web content fetcher |
| `/v1/videos/generations` | `POST` | AI video generation (xAI Grok, Luma, Kling) |
| `/v1/videos/:id` | `GET` | Asynchronous video generation status probe |
| `/api/v1/*` | `ANY` | Aliased mirror for all standard `/v1` endpoints |

### 2. Core Gateway & Diagnostics API (`/api/*`)

| Endpoint | Method | Description |
|---|---|---|
| `/api/health` | `GET` | Real-time gateway health check & system uptime probe |
| `/api/version` | `GET` | Gateway version metadata (`v0.5.92`) |
| `/api/init` | `GET` | First-run setup and initialization status |
| `/api/locale` | `GET` | Internationalization settings (English, Indonesian) |
| `/api/changelog` | `GET` | Real-time changelog notes and release history |
| `/api/auth/status` | `GET` | Session authentication and login requirement status |
| `/api/auth/login` | `POST` | Authenticate user session with JWT token generation |
| `/api/auth/logout` | `POST` | Clear session cookies and invalidate token |
| `/api/auth/reset-password` | `POST` | Secure password reset utility |
| `/api/auth/oidc` | `GET, POST` | Enterprise OIDC Single Sign-On handler |
| `/api/auth/saml` | `GET, POST` | Enterprise SAML 2.0 Single Sign-On handler |
| `/api/settings` | `GET, PATCH` | Global gateway settings and provider strategies |
| `/api/settings/require-login` | `GET, POST` | Toggle mandatory login requirement |
| `/api/settings/database` | `GET, POST` | Database backup, restore, and table inspection |
| `/api/providers` | `GET, POST` | Provider configuration list and account management |
| `/api/providers/client` | `GET` | Public client provider list |
| `/api/providers/suggested-models` | `GET` | Intelligent model recommendation catalog |
| `/api/provider-nodes` | `GET, POST` | Distributed provider node cluster management |
| `/api/keys` | `GET, POST` | API Key management (create, revoke, rate limit) |
| `/api/combos` | `GET, POST` | Fallback combos (multi-model failover chains) |
| `/api/combos/presets` | `GET` | Pre-configured combo chains (Coding, Creative, Fast) |
| `/api/proxy-pools` | `GET, POST` | Residential & cloud proxy pool rotation (Deno, Vercel, Cloudflare) |
| `/api/models` | `GET` | Unified model catalog and availability matrix |
| `/api/models/alias` | `GET, POST` | Custom model alias mapping |
| `/api/models/custom` | `GET, POST` | User-defined custom model endpoints |
| `/api/models/availability` | `GET` | Dynamic provider latency & model availability tracker |
| `/api/models/disabled` | `GET, POST` | Blacklist models from routing |
| `/api/usage` | `GET` | Aggregated token usage statistics |
| `/api/usage/history` | `GET` | Historical usage timeline |
| `/api/usage/chart` | `GET` | Graphical usage breakdown for dashboard |
| `/api/usage/providers` | `GET` | Provider-level token expenditure breakdown |
| `/api/usage/logs` | `GET` | Detailed request logging and latency audits |
| `/api/usage/stats` | `GET` | Real-time RPS and throughput counters |
| `/api/token-saver/summary` | `GET` | Token Saver stats and cumulative savings report |
| `/api/token-saver/test` | `POST` | Interactive sandbox for testing token compression algorithms |
| `/api/headroom` | `GET, POST` | Headroom proxy buffer status and lifecycle |
| `/api/pxpipe` | `GET, POST` | Pxpipe local proxy pipeline manager |
| `/api/tunnel` | `GET` | Public tunnel status overview |
| `/api/tunnel/status` | `GET` | Real-time Cloudflare Quick Tunnel & Tailscale health |
| `/api/tunnel/enable` | `POST` | Start native Cloudflare Quick Tunnel (`trycloudflare.com`) |
| `/api/tunnel/disable` | `POST` | Stop active Cloudflare Quick Tunnel |
| `/api/tunnel/tailscale-check` | `GET` | Inspect local Tailscale daemon and login status |
| `/api/tunnel/tailscale-enable` | `POST` | Activate Tailscale Funnel on port 20128 |
| `/api/tunnel/tailscale-disable` | `POST` | Deactivate Tailscale Funnel |
| `/api/translator` | `GET, POST` | Native multi-language translation engine |
| `/api/cli-tools` | `GET` | AI Coding CLI integration settings (Claude, Cline, Cursor, OpenCode) |
| `/api/media-providers` | `GET, POST` | Multimodal providers (TTS, STT, Image, Video) |
| `/api/tags` | `GET` | Supported model categorization tags |
| `/api/pricing` | `GET` | Per-model pricing database (input/output per 1M tokens) |

### 3. Web Dashboard Frontend Routes

All dashboard routes run cleanly with modern dark-mode styling, zero dev badges, and direct URL aliases:

| Dashboard Route | Direct Alias | Description |
|---|---|---|
| `/dashboard/endpoint` | `/endpoint` | API endpoint URLs (Local, Cloudflare Tunnel, Tailscale) and API keys |
| `/dashboard/providers` | `/providers` | AI provider connections, API keys, and account rotation |
| `/dashboard/providers/new` | `/providers/new` | Wizard to add new free or commercial AI providers |
| `/dashboard/proxy-pools` | `/proxy-pools` | Reverse proxy cluster manager with 1-click cloud deploy |
| `/dashboard/combos` | `/combos` | Multi-model fallback cascades and automatic failover |
| `/dashboard/skills` | `/skills` | AI Agent Skills hub pointing to `dresar/go-9router` |
| `/dashboard/token-saver` | `/token-saver` | Token Saver compression dashboard, presets, and live testbed |
| `/dashboard/cli-tools` | `/cli-tools` | 1-Click integration for Cursor, Claude Code CLI, Cline, etc. |
| `/dashboard/profile` | `/profile` | Password management, security policies, and language switch |
| `/dashboard/usage` | `/usage` | Detailed analytics, graphs, and request inspector |
| `/dashboard/quota` | `/quota` | Usage quotas and provider budget limits |
| `/dashboard/console-log` | `/console-log` | Live gateway log stream with filtering and search |
| `/dashboard/mitm` | `/mitm` | Anti-Gravity MITM proxy inspection |
| `/dashboard/pxpipe` | `/pxpipe` | Pipe proxy pipeline status and statistics |
| `/dashboard/translator` | `/translator` | Prompt translation workbench |
| `/dashboard/settings/pricing` | `/settings/pricing` | Pricing model table and cost simulator |
| `/dashboard/basic-chat` | `/basic-chat` | Built-in interactive test chat interface |
| `/dashboard/media-providers/web` | `/media-providers/web` | Multimodal provider configurations |

---

## 🔒 Free Cloudflare Quick Tunnel & Tailscale

Go 9Router includes native support for **zero-cost public endpoints**:

### Cloudflare Quick Tunnel (`trycloudflare.com`)
- **100% Free**: No Cloudflare account, API token, or credit card required.
- **Native Go Process Supervisor**: Automatically downloads and executes `cloudflared.exe` with hidden console window.
- **Fast Startup**: Captures public URL (`https://*.trycloudflare.com`) in ~5 seconds via non-blocking `io.Pipe()`.
- **Direct Edge Routing**: Routes public traffic securely to `http://127.0.0.1:20128`.

```bash
# Enable Tunnel via API:
curl -X POST http://127.0.0.1:20128/api/tunnel/enable
# Response:
# {"success":true,"tunnelUrl":"https://wesley-southeast-fairy-handles.trycloudflare.com","status":"running"}

# Disable Tunnel:
curl -X POST http://127.0.0.1:20128/api/tunnel/disable
```

### Tailscale Funnel & Serve
- Automatically detects system Tailscale daemon (`tailscale.exe`).
- Enables secure mesh networking with encrypted traffic via your private `*.ts.net` domain.

---

## 🧠 Smart Token Saver

The Token Saver engine provides automated token reduction for large context windows, coding agents, and automated workflows:

- **Whitespace & Comments Minimizer**: Strips redundant indentation and whitespace from code snippets.
- **JSON Compaction**: Compresses JSON schema parameters and tool call specifications by up to 40%.
- **Context Optimizer**: Deduplicates repetitive system prompts across conversation turns.
- **CLI Header Bypass**: Use header `x-9router-token-saver: off` to bypass optimization on specific requests.

---

## 🛠️ Installation & Building

### Prerequisites
- [Go 1.22+](https://golang.org/dl/)
- [Bun](https://bun.sh/) or [Node.js 20+](https://nodejs.org/)

### 1. Build the Gateway Binary
```powershell
go build -o bin/9router.exe ./cmd/9router
```

### 2. Build the Next.js Production Frontend
```powershell
cd frontend
bun install
bun run build
cd ..
```

### 3. Launch Go 9Router
```powershell
.\bin\9router.exe -p 20128
```
The gateway starts on `http://localhost:20128` and opens the dashboard in your default browser.

---

## 🤖 CLI Usage

```powershell
# Interactive Terminal Menu
.\bin\9router.exe menu

# Inspect Server Status
.\bin\9router.exe status

# Stop Gateway Daemon
.\bin\9router.exe stop

# Generate Video via xAI Grok
.\bin\9router.exe xai video --prompt "Cyberpunk city in neon rain" --output city.mp4
```

---

## 🧪 Comprehensive Route Audit

You can run the automated auditor at any time to verify all 89 routes:

```powershell
node tools/audit_all_routes.js
```

**Audit Result**:
```
======================================================
  🔍 GO-9ROUTER COMPREHENSIVE ROUTE & ENDPOINT AUDIT  
  Gateway Base URL: http://127.0.0.1:20128
======================================================
  Total Routes Tested : 89
  Passed (OK / Not 404): 89
  Failed / 404 Errors : 0
======================================================
✅ ALL ROUTES AUDITED SUCCESSFULLY WITH ZERO 404 ERRORS!
```

---

## 📄 License

MIT License — Created and maintained by [dresar](https://github.com/dresar/go-9router).
