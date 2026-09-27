package tunnel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Manager struct {
	mu          sync.RWMutex
	cmd         *exec.Cmd
	running     bool
	tunnelURL   string
	startedAt   time.Time
	lastError   string
	cancelFunc  context.CancelFunc
	logs        []string
	tsURL       string
	tsRunning   bool
	tsInstalled bool
}

var (
	DefaultManager = &Manager{
		logs: make([]string, 0, 100),
	}
	urlRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
)

func FindCloudflared() string {
	candidates := []string{
		filepath.Join(os.Getenv("APPDATA"), "9router", "bin", "cloudflared.exe"),
		filepath.Join(".", "data", "bin", "cloudflared.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "cloudflared", "cloudflared.exe"),
		`C:\Program Files\cloudflared\cloudflared.exe`,
		`C:\Program Files (x86)\cloudflared\cloudflared.exe`,
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() && fi.Size() > 10*1024*1024 {
			return c
		}
	}

	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p
	}
	if p, err := exec.LookPath("cloudflared.exe"); err == nil {
		return p
	}

	return ""
}

func (m *Manager) IsCloudflaredDownloading() (bool, int) {
	candidates := []string{
		filepath.Join(os.Getenv("APPDATA"), "9router", "bin", "cloudflared.exe"),
		filepath.Join(".", "data", "bin", "cloudflared.exe"),
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			if fi.Size() > 0 && fi.Size() < 50*1024*1024 {
				pct := int((float64(fi.Size()) / (53.0 * 1024 * 1024)) * 100)
				if pct > 99 {
					pct = 99
				}
				return true, pct
			}
		}
	}
	return false, 0
}

func (m *Manager) StartCloudflare(port int) (string, error) {
	m.mu.Lock()
	if m.running && m.tunnelURL != "" {
		url := m.tunnelURL
		m.mu.Unlock()
		return url, nil
	}
	m.mu.Unlock()

	binPath := FindCloudflared()
	if binPath == "" {
		if dl, pct := m.IsCloudflaredDownloading(); dl {
			return "", fmt.Errorf("Cloudflare tunnel binary is currently downloading (%d%%). Please wait a few moments and try again.", pct)
		}
		return "", fmt.Errorf("Cloudflare binary (cloudflared.exe) not found. Downloading in progress, please try again shortly.")
	}

	m.mu.Lock()
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel

	cmd := exec.CommandContext(ctx, binPath, "tunnel", "--url", fmt.Sprintf("http://127.0.0.1:%d", port), "--no-autoupdate", "--retries", "99")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000,
		}
	}

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	m.cmd = cmd
	m.running = false
	m.tunnelURL = ""
	m.lastError = ""
	m.startedAt = time.Now()
	m.mu.Unlock()

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		m.mu.Lock()
		m.lastError = err.Error()
		m.mu.Unlock()
		return "", fmt.Errorf("failed to start cloudflared: %w", err)
	}

	readyChan := make(chan string, 1)

	go func() {
		scanner := bufio.NewScanner(pr)
		urlFound := false
		for scanner.Scan() {
			line := scanner.Text()
			m.mu.Lock()
			m.logs = append(m.logs, line)
			if len(m.logs) > 100 {
				m.logs = m.logs[1:]
			}
			m.mu.Unlock()

			if !urlFound {
				if match := urlRegex.FindString(line); match != "" && !strings.Contains(match, "api.trycloudflare.com") {
					urlFound = true
					m.mu.Lock()
					m.running = true
					m.tunnelURL = match
					m.mu.Unlock()
					readyChan <- match
				}
			}
		}

		_ = cmd.Wait()
		pw.Close()
		m.mu.Lock()
		m.running = false
		m.tunnelURL = ""
		m.mu.Unlock()
	}()

	select {
	case url := <-readyChan:
		return url, nil
	case <-time.After(45 * time.Second):
		m.StopCloudflare()
		return "", fmt.Errorf("timeout waiting for Cloudflare Quick Tunnel URL. Please check network connectivity.")
	}
}

func (m *Manager) StopCloudflare() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancelFunc != nil {
		m.cancelFunc()
		m.cancelFunc = nil
	}
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		m.cmd = nil
	}
	m.running = false
	m.tunnelURL = ""
	return nil
}

func (m *Manager) GetStatus() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dl, pct := m.IsCloudflaredDownloading()
	var downloadInfo any = nil
	if dl {
		downloadInfo = map[string]any{
			"downloading": true,
			"progress":    pct,
		}
	}

	tsStatus := m.CheckTailscale()

	statusStr := "stopped"
	if m.running {
		statusStr = "running"
	}

	return map[string]any{
		"tunnel": map[string]any{
			"enabled":         m.running,
			"settingsEnabled": m.running,
			"status":          statusStr,
			"running":         m.running,
			"tunnelUrl":       m.tunnelURL,
			"publicUrl":       m.tunnelURL,
		},
		"tailscale": tsStatus,
		"download":  downloadInfo,
	}
}

func FindTailscale() string {
	candidates := []string{
		`C:\Program Files\Tailscale\tailscale.exe`,
		`C:\Program Files (x86)\Tailscale\tailscale.exe`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Tailscale", "tailscale.exe"),
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	if p, err := exec.LookPath("tailscale"); err == nil {
		return p
	}
	if p, err := exec.LookPath("tailscale.exe"); err == nil {
		return p
	}
	return ""
}

func (m *Manager) CheckTailscale() map[string]any {
	tsBin := FindTailscale()
	if tsBin == "" {
		return map[string]any{
			"installed":           false,
			"loggedIn":            false,
			"platform":            runtime.GOOS,
			"brewAvailable":       false,
			"daemonRunning":       false,
			"customDaemonRunning": false,
			"systemDaemonRunning": false,
			"hasCachedPassword":   false,
			"enabled":             false,
			"status":              "stopped",
			"url":                 "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, tsBin, "status", "--json")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	}
	out, err := cmd.Output()
	if err != nil {
		return map[string]any{
			"installed":     true,
			"loggedIn":      false,
			"platform":      runtime.GOOS,
			"daemonRunning": false,
			"enabled":       false,
			"status":        "stopped",
			"url":           "",
		}
	}

	var parsed struct {
		BackendState string `json:"BackendState"`
		Self         struct {
			DNSName string `json:"DNSName"`
		} `json:"Self"`
	}
	_ = json.Unmarshal(out, &parsed)

	isLoggedIn := parsed.BackendState == "Running"
	dnsName := strings.TrimSuffix(parsed.Self.DNSName, ".")
	var url string
	if dnsName != "" {
		url = fmt.Sprintf("https://%s:20128", dnsName)
	}

	return map[string]any{
		"installed":           true,
		"loggedIn":            isLoggedIn,
		"platform":            runtime.GOOS,
		"daemonRunning":       isLoggedIn,
		"customDaemonRunning": false,
		"systemDaemonRunning": isLoggedIn,
		"hasCachedPassword":   false,
		"enabled":             m.tsRunning && isLoggedIn,
		"status":              ternary(m.tsRunning && isLoggedIn, "running", "stopped"),
		"tunnelUrl":           url,
		"url":                 url,
	}
}

func (m *Manager) EnableTailscale(port int) (string, error) {
	tsBin := FindTailscale()
	if tsBin == "" {
		return "", fmt.Errorf("Tailscale is not installed on this system. Download from https://tailscale.com/download/windows")
	}

	status := m.CheckTailscale()
	if loggedIn, ok := status["loggedIn"].(bool); !ok || !loggedIn {
		return "", fmt.Errorf("Tailscale is installed but not logged in. Please log in to your Tailscale app.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, tsBin, "funnel", fmt.Sprintf("%d", port), "on")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	}
	_ = cmd.Run()

	m.mu.Lock()
	m.tsRunning = true
	m.mu.Unlock()

	updated := m.CheckTailscale()
	if u, ok := updated["tunnelUrl"].(string); ok && u != "" {
		return u, nil
	}
	return "", nil
}

func (m *Manager) DisableTailscale(port int) error {
	tsBin := FindTailscale()
	if tsBin != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, tsBin, "funnel", fmt.Sprintf("%d", port), "off")
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		}
		_ = cmd.Run()
	}
	m.mu.Lock()
	m.tsRunning = false
	m.mu.Unlock()
	return nil
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
