package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/dresar/go-9router/internal/logging"
)

type Status struct {
	CurrentVersion   string    `json:"currentVersion"`
	CurrentCommit    string    `json:"currentCommit"`
	CurrentCommitMsg string    `json:"currentCommitMsg"`
	LatestCommit     string    `json:"latestCommit"`
	LatestCommitMsg  string    `json:"latestCommitMsg"`
	LatestCommitDate string    `json:"latestCommitDate"`
	HasUpdate        bool      `json:"hasUpdate"`
	LastChecked      time.Time `json:"lastChecked"`
	IsChecking       bool      `json:"isChecking"`
	IsSyncing        bool      `json:"isSyncing"`
	LastError        string    `json:"lastError,omitempty"`
	RepoURL          string    `json:"repoUrl"`
	IntervalSeconds  int       `json:"intervalSeconds"`
}

type Manager struct {
	mu         sync.RWMutex
	status     Status
	repoOwner  string
	repoName   string
	branch     string
	ticker     *time.Ticker
	stopCh     chan struct{}
	checkMutex sync.Mutex
	syncMutex  sync.Mutex
}

var (
	defaultManager *Manager
	once           sync.Once
)

func GetManager() *Manager {
	once.Do(func() {
		defaultManager = &Manager{
			repoOwner: "dresar",
			repoName:  "go-9router",
			branch:    "master",
			stopCh:    make(chan struct{}),
			status: Status{
				CurrentVersion:  "0.1.0",
				RepoURL:         "https://github.com/dresar/go-9router",
				IntervalSeconds: 3600,
			},
		}
		defaultManager.initLocalCommit()
	})
	return defaultManager
}

func (m *Manager) initLocalCommit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	if out, err := cmd.Output(); err == nil {
		m.status.CurrentCommit = strings.TrimSpace(string(out))
	}
	msgCmd := exec.Command("git", "log", "-1", "--pretty=%s")
	if out, err := msgCmd.Output(); err == nil {
		m.status.CurrentCommitMsg = strings.TrimSpace(string(out))
	}
}

func (m *Manager) GetStatus() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *Manager) CheckNow(ctx context.Context) (Status, error) {
	m.checkMutex.Lock()
	defer m.checkMutex.Unlock()

	m.mu.Lock()
	m.status.IsChecking = true
	m.status.LastError = ""
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.status.IsChecking = false
		m.status.LastChecked = time.Now()
		m.mu.Unlock()
	}()

	// 1. Refresh local commit
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
	if out, err := cmd.Output(); err == nil {
		m.mu.Lock()
		m.status.CurrentCommit = strings.TrimSpace(string(out))
		m.mu.Unlock()
	}
	msgCmd := exec.CommandContext(ctx, "git", "log", "-1", "--pretty=%s")
	if out, err := msgCmd.Output(); err == nil {
		m.mu.Lock()
		m.status.CurrentCommitMsg = strings.TrimSpace(string(out))
		m.mu.Unlock()
	}

	// 2. Query GitHub Commits API for latest commit info
	var latestSHA, latestMsg, latestDate string
	var apiErr error

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", m.repoOwner, m.repoName, m.branch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "Go-9Router-Updater")
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var result struct {
					SHA    string `json:"sha"`
					Commit struct {
						Message   string `json:"message"`
						Committer struct {
							Date string `json:"date"`
						} `json:"committer"`
					} `json:"commit"`
				}
				if json.NewDecoder(resp.Body).Decode(&result) == nil && result.SHA != "" {
					if len(result.SHA) >= 7 {
						latestSHA = result.SHA[:7]
					} else {
						latestSHA = result.SHA
					}
					lines := strings.Split(result.Commit.Message, "\n")
					if len(lines) > 0 {
						latestMsg = strings.TrimSpace(lines[0])
					}
					latestDate = result.Commit.Committer.Date
				}
			} else {
				apiErr = fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
			}
		} else {
			apiErr = err
		}
	}

	// Fallback to git ls-remote if GitHub API was unavailable
	if latestSHA == "" {
		lsCmd := exec.CommandContext(ctx, "git", "ls-remote", "origin", fmt.Sprintf("refs/heads/%s", m.branch))
		if out, err := lsCmd.Output(); err == nil {
			parts := strings.Fields(string(out))
			if len(parts) > 0 && len(parts[0]) >= 7 {
				latestSHA = parts[0][:7]
			}
		} else if apiErr != nil {
			m.mu.Lock()
			m.status.LastError = fmt.Sprintf("Gagal cek update: %v", apiErr)
			m.mu.Unlock()
			return m.GetStatus(), apiErr
		}
	}

	m.mu.Lock()
	if latestSHA != "" {
		m.status.LatestCommit = latestSHA
		if latestMsg != "" {
			m.status.LatestCommitMsg = latestMsg
		}
		if latestDate != "" {
			m.status.LatestCommitDate = latestDate
		}
		if m.status.CurrentCommit != "" && latestSHA != "" {
			m.status.HasUpdate = (m.status.CurrentCommit != latestSHA)
		}
	}
	curr := m.status.CurrentCommit
	late := m.status.LatestCommit
	hasUp := m.status.HasUpdate
	m.mu.Unlock()

	logging.Info("UPDATER", fmt.Sprintf("sync check completed: local=%s, remote=%s, hasUpdate=%v", curr, late, hasUp))
	return m.GetStatus(), nil
}

func (m *Manager) SyncNow(ctx context.Context) (string, error) {
	m.syncMutex.Lock()
	defer m.syncMutex.Unlock()

	m.mu.Lock()
	m.status.IsSyncing = true
	m.status.LastError = ""
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.status.IsSyncing = false
		m.mu.Unlock()
	}()

	cmd := exec.CommandContext(ctx, "git", "pull", "origin", m.branch)
	out, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(out))

	if err != nil {
		m.mu.Lock()
		m.status.LastError = fmt.Sprintf("Git pull failed: %v (%s)", err, outputStr)
		m.mu.Unlock()
		logging.Warn("UPDATER", "sync failed", "err", err, "output", outputStr)
		return outputStr, err
	}

	// Update local commit info
	m.initLocalCommit()

	m.mu.Lock()
	m.status.HasUpdate = false
	m.status.LatestCommit = m.status.CurrentCommit
	m.status.LatestCommitMsg = m.status.CurrentCommitMsg
	m.status.LastError = ""
	m.mu.Unlock()

	logging.Info("UPDATER", fmt.Sprintf("sync successful: %s", outputStr))
	return outputStr, nil
}

func (m *Manager) StartScheduler(interval time.Duration) {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	m.mu.Lock()
	m.status.IntervalSeconds = int(interval.Seconds())
	m.mu.Unlock()

	m.ticker = time.NewTicker(interval)

	go func() {
		// Initial check 5s after startup
		time.Sleep(5 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		_, _ = m.CheckNow(ctx)
		cancel()

		for {
			select {
			case <-m.ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				logging.Info("UPDATER", "running scheduled 1-hour GitHub sync check...")
				_, _ = m.CheckNow(ctx)
				cancel()
			case <-m.stopCh:
				m.ticker.Stop()
				return
			}
		}
	}()
}

func (m *Manager) Stop() {
	if m.stopCh != nil {
		close(m.stopCh)
	}
}
