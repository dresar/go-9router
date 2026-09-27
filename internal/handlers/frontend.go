package handlers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (h *Handler) HandleFrontend(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") {
		http.NotFound(w, r)
		return
	}

	// 1. Check if static exported build exists in frontend/out or web/out
	for _, dir := range []string{"frontend/out", "web/out", "public"} {
		filePath := filepath.Join(dir, strings.TrimPrefix(r.URL.Path, "/"))
		if fi, err := os.Stat(filePath); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, filePath)
			return
		}
		indexPath := filepath.Join(filePath, "index.html")
		if fi, err := os.Stat(indexPath); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, indexPath)
			return
		}
	}

	// 2. Check if Next.js dev server is running on port 20127
	if isPortOpen("127.0.0.1", 20127) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			proxyWebSocket(w, r, "127.0.0.1:20127")
			return
		}
		target, _ := url.Parse("http://127.0.0.1:20127")
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
		return
	}

	// 3. Fallback: Modern, rich dark-mode built-in Web Dashboard
	serveEmbeddedDashboard(w, r, h.Cfg.Port)
}

func proxyWebSocket(w http.ResponseWriter, r *http.Request, targetAddr string) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server does not support hijacking", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	upstreamConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	defer upstreamConn.Close()

	if err := r.Write(upstreamConn); err != nil {
		return
	}

	errc := make(chan error, 2)
	go func() {
		_, err := io.Copy(upstreamConn, clientConn)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(clientConn, upstreamConn)
		errc <- err
	}()
	<-errc
}

func isPortOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 100*time.Millisecond)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func serveEmbeddedDashboard(w http.ResponseWriter, r *http.Request, port string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Go 9Router - AI Gateway Dashboard</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #090a0f;
      --card-bg: #12141c;
      --card-border: #1e2230;
      --text: #e2e8f0;
      --text-muted: #94a3b8;
      --accent: #38bdf8;
      --accent-glow: rgba(56, 189, 248, 0.15);
      --green: #22c55e;
      --purple: #a855f7;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: 'Inter', sans-serif;
      background-color: var(--bg);
      color: var(--text);
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }
    header {
      border-bottom: 1px solid var(--card-border);
      background: rgba(18, 20, 28, 0.8);
      backdrop-filter: blur(12px);
      padding: 1rem 2rem;
      display: flex;
      align-items: center;
      justify-content: space-between;
      position: sticky;
      top: 0;
      z-index: 50;
    }
    .logo {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      font-weight: 700;
      font-size: 1.25rem;
      letter-spacing: -0.02em;
    }
    .badge {
      background: rgba(34, 197, 94, 0.15);
      color: var(--green);
      border: 1px solid rgba(34, 197, 94, 0.3);
      font-size: 0.75rem;
      padding: 0.2rem 0.5rem;
      border-radius: 9999px;
      font-weight: 600;
      display: flex;
      align-items: center;
      gap: 0.35rem;
    }
    .badge::before {
      content: '';
      width: 6px;
      height: 6px;
      background: var(--green);
      border-radius: 50%%;
      box-shadow: 0 0 8px var(--green);
    }
    main {
      flex: 1;
      max-width: 1200px;
      margin: 0 auto;
      padding: 2.5rem 1.5rem;
      width: 100%%;
    }
    .hero {
      margin-bottom: 2.5rem;
    }
    .hero h1 {
      font-size: 2.25rem;
      font-weight: 800;
      letter-spacing: -0.03em;
      margin-bottom: 0.5rem;
      background: linear-gradient(135deg, #fff 0%%, #94a3b8 100%%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }
    .hero p {
      color: var(--text-muted);
      font-size: 1.05rem;
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
      gap: 1.25rem;
      margin-bottom: 2rem;
    }
    .card {
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 12px;
      padding: 1.5rem;
      transition: border-color 0.2s, box-shadow 0.2s;
    }
    .card:hover {
      border-color: rgba(56, 189, 248, 0.4);
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
    }
    .card h3 {
      font-size: 1rem;
      font-weight: 600;
      margin-bottom: 0.5rem;
      color: #fff;
    }
    .card p {
      color: var(--text-muted);
      font-size: 0.875rem;
      line-height: 1.5;
    }
    .code-box {
      background: #05070a;
      border: 1px solid #1a1e2d;
      border-radius: 8px;
      padding: 1rem;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.85rem;
      color: #38bdf8;
      overflow-x: auto;
      margin-top: 1rem;
    }
    .btn {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      background: #0284c7;
      color: #fff;
      padding: 0.6rem 1.2rem;
      border-radius: 8px;
      text-decoration: none;
      font-size: 0.875rem;
      font-weight: 500;
      margin-top: 1rem;
      transition: background 0.15s;
    }
    .btn:hover { background: #0369a1; }
    footer {
      border-top: 1px solid var(--card-border);
      padding: 1.5rem 2rem;
      text-align: center;
      color: var(--text-muted);
      font-size: 0.875rem;
    }
  </style>
</head>
<body>
  <header>
    <div class="logo">
      <span>⚡ Go 9Router</span>
    </div>
    <div style="display: flex; gap: 1rem; align-items: center;">
      <span class="badge">Online (Port %s)</span>
      <a href="/api/health" class="btn" style="margin: 0; padding: 0.4rem 0.8rem; font-size: 0.75rem;">Health API</a>
    </div>
  </header>

  <main>
    <div class="hero">
      <h1>Unified AI Gateway & Proxy</h1>
      <p>High-performance Golang backend serving OpenAI, Anthropic, Gemini, Grok, and DeepSeek endpoints.</p>
    </div>

    <div class="grid">
      <div class="card">
        <h3>🚀 OpenAI & Claude Proxy</h3>
        <p>Drop-in replacement for OpenAI SDK, Cursor, Claude Code, and Aider.</p>
        <div class="code-box">http://localhost:%s/v1</div>
      </div>

      <div class="card">
        <h3>🔑 Provider Management</h3>
        <p>Active accounts, failover pools, and custom model routing stored in SQLite.</p>
        <div class="code-box">GET /api/providers</div>
      </div>

      <div class="card">
        <h3>📊 Realtime Usage & SSE</h3>
        <p>Fast streaming token tracking and SSE metrics directly in Go.</p>
        <div class="code-box">GET /api/usage/stream</div>
      </div>
    </div>

    <div class="card">
      <h3>Quick Verification</h3>
      <p>Test the gateway directly using curl from your terminal:</p>
      <div class="code-box">curl -X POST http://localhost:%s/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello!"}]}'</div>
    </div>
  </main>

  <footer>
    Go 9Router • Built for high concurrency, ultra-low memory, and instant startup.
  </footer>
</body>
</html>`, port, port, port)
	w.Write([]byte(html))
}
