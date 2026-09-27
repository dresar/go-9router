package handlers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	nextServerMu  sync.Mutex
	nextServerCmd *exec.Cmd
)

func EnsureNextServer() {
	nextServerMu.Lock()
	defer nextServerMu.Unlock()

	if isPortOpen("127.0.0.1", 20127) {
		return
	}

	standaloneDir := filepath.Join("frontend", ".next", "standalone")
	serverFile := filepath.Join(standaloneDir, "custom-server.js")
	if _, err := os.Stat(serverFile); err != nil {
		serverFile = filepath.Join(standaloneDir, "server.js")
		if _, err := os.Stat(serverFile); err != nil {
			return
		}
	}

	nodeBin, err := exec.LookPath("node")
	if err != nil {
		return
	}

	cmd := exec.Command(nodeBin, filepath.Base(serverFile))
	cmd.Dir = standaloneDir
	cmd.Env = append(os.Environ(), "PORT=20127", "HOSTNAME=0.0.0.0", "NODE_ENV=production")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000,
		}
	}

	if err := cmd.Start(); err == nil {
		nextServerCmd = cmd
		for i := 0; i < 30; i++ {
			time.Sleep(100 * time.Millisecond)
			if isPortOpen("127.0.0.1", 20127) {
				break
			}
		}
	}
}

func (h *Handler) HandleFrontend(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/v1/") {
		http.NotFound(w, r)
		return
	}

	if r.URL.Path == "/" || r.URL.Path == "" || r.URL.Path == "/dashboard" {
		http.Redirect(w, r, "/endpoint", http.StatusTemporaryRedirect)
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

	// 2. Ensure Next.js dashboard server on port 20127 is active
	if !isPortOpen("127.0.0.1", 20127) {
		EnsureNextServer()
	}

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

	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	http.Error(w, "Next.js dashboard server is initializing or unavailable on port 20127", http.StatusServiceUnavailable)
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

func (h *Handler) HandleMCPProxy(w http.ResponseWriter, r *http.Request) {
	if !isPortOpen("127.0.0.1", 20127) {
		EnsureNextServer()
	}

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
	h.JSONError(w, http.StatusServiceUnavailable, "Next.js backend with MCP stdio bridge is not running on port 20127")
}
