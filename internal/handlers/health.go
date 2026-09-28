package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strings"

	"github.com/dresar/go-9router/internal/updater"
)

const version = "0.1.0"

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) HandleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	mgr := updater.GetManager()
	st := mgr.GetStatus()

	h.JSON(w, http.StatusOK, map[string]any{
		"version":          version,
		"currentVersion":   version,
		"latestVersion":    st.LatestCommit,
		"hasUpdate":        st.HasUpdate,
		"currentCommit":    st.CurrentCommit,
		"currentCommitMsg": st.CurrentCommitMsg,
		"latestCommit":     st.LatestCommit,
		"latestCommitMsg":  st.LatestCommitMsg,
		"latestCommitDate": st.LatestCommitDate,
		"lastChecked":      st.LastChecked,
		"isChecking":       st.IsChecking,
		"isSyncing":        st.IsSyncing,
		"lastError":        st.LastError,
		"repoUrl":          st.RepoURL,
		"intervalSeconds":  st.IntervalSeconds,
		"go":               runtime.Version(),
		"os":               runtime.GOOS,
		"arch":             runtime.GOARCH,
	})
}

func (h *Handler) HandleVersionCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	mgr := updater.GetManager()
	st, err := mgr.CheckNow(r.Context())
	if err != nil {
		h.JSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
			"status":  st,
		})
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"status":  st,
	})
}

func (h *Handler) HandleVersionSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	mgr := updater.GetManager()
	output, err := mgr.SyncNow(r.Context())
	if err != nil {
		h.JSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
			"output":  output,
		})
		return
	}
	st := mgr.GetStatus()
	h.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"output":  output,
		"status":  st,
	})
}

func (h *Handler) HandleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"initialized": true,
		"version":     version,
	})
}

func (h *Handler) HandleLocale(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		loc := "en"
		if c, err := r.Cookie("locale"); err == nil && c.Value != "" {
			loc = c.Value
		}
		h.JSON(w, http.StatusOK, map[string]any{"locale": loc})
	case http.MethodPost:
		var body struct {
			Locale string `json:"locale"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		loc := strings.TrimSpace(body.Locale)
		if loc != "id" && loc != "en" {
			loc = "en"
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "locale",
			Value:    loc,
			Path:     "/",
			MaxAge:   86400 * 365,
			SameSite: http.SameSiteLaxMode,
		})
		h.JSON(w, http.StatusOK, map[string]any{"success": true, "locale": loc})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	go func() {
		_ = r.Context().Err()
	}()
}
