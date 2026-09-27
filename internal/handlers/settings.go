package handlers

import (
	"net/http"

	"github.com/dresar/go-9router/internal/auth"
	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getSettings(w, r)
	case http.MethodPatch:
		h.updateSettings(w, r)
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := repos.GetSettings(h.DB)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get settings")
		return
	}
	safe := make(repos.Settings)
	for k, v := range settings {
		safe[k] = v
	}
	pwd, _ := safe["password"].(string)
	oidcSecret, _ := safe["oidcClientSecret"].(string)
	delete(safe, "password")
	delete(safe, "oidcClientSecret")
	safe["hasPassword"] = pwd != ""
	safe["oidcConfigured"] = oidcSecret != "" &&
		safe["oidcIssuerUrl"] != nil && safe["oidcClientId"] != nil
	safe["enableRequestLogs"] = h.Cfg.EnableRequestLogs
	safe["enableTranslator"] = false
	w.Header().Set("Cache-Control", "no-store")
	h.JSON(w, http.StatusOK, safe)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := h.DecodeJSON(r, &body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	protected := []string{"password", "mitmSudoEncrypted"}
	for _, k := range protected {
		delete(body, k)
	}

	if newPwd, ok := body["newPassword"].(string); ok && newPwd != "" {
		settings, _ := repos.GetSettings(h.DB)
		currentHash := repos.SettingStr(settings, "password", "")
		if currentHash != "" {
			currentPwd, _ := body["currentPassword"].(string)
			if currentPwd == "" {
				h.JSONError(w, http.StatusBadRequest, "Current password required")
				return
			}
			if !auth.CheckPassword(currentHash, currentPwd) {
				h.JSONError(w, http.StatusUnauthorized, "Invalid current password")
				return
			}
		}
		hash, err := auth.HashPassword(newPwd)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "password hash error")
			return
		}
		body["password"] = hash
		delete(body, "newPassword")
		delete(body, "currentPassword")
	}

	if v, ok := body["oidcClientSecret"]; ok {
		if v == nil || v == "" {
			delete(body, "oidcClientSecret")
		}
	}

	settings, err := repos.UpdateSettings(h.DB, body)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}
	safe := make(repos.Settings)
	for k, v := range settings {
		safe[k] = v
	}
	oidcSecret, _ := safe["oidcClientSecret"].(string)
	delete(safe, "password")
	delete(safe, "oidcClientSecret")
	safe["oidcConfigured"] = oidcSecret != "" &&
		safe["oidcIssuerUrl"] != nil && safe["oidcClientId"] != nil
	w.Header().Set("Cache-Control", "no-store")
	h.JSON(w, http.StatusOK, safe)
}
