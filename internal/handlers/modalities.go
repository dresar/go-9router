package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/providers"
	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if body["input"] == nil {
		h.JSONError(w, http.StatusBadRequest, "Missing required field: input")
		return
	}

	modelStr, _ := body["model"].(string)
	if modelStr == "" {
		modelStr = "openai/text-embedding-3-small"
	}

	providerID, modelID, ok := providers.ResolveModelProvider(modelStr, h.DB)
	if !ok {
		providerID = "openai"
		modelID = modelStr
	}

	sel, err := providers.SelectCredentials(h.DB, providerID, nil)
	if err != nil || sel == nil {
		h.JSONError(w, http.StatusNotFound, fmt.Sprintf("no active credentials for provider: %s", providerID))
		return
	}

	body["model"] = modelID
	data, _ := json.Marshal(body)

	baseURL := sel.Credentials.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, baseURL+"/embeddings", bytes.NewReader(data))
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to build request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if sel.Credentials.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+sel.Credentials.APIKey)
	}

	client := &http.Client{Timeout: providers.DefaultClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, "upstream error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) HandleImagesGenerations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	prompt, _ := body["prompt"].(string)
	if strings.TrimSpace(prompt) == "" {
		h.JSONError(w, http.StatusBadRequest, "Missing required field: prompt")
		return
	}

	modelStr, _ := body["model"].(string)
	if modelStr == "" {
		modelStr = "openai/dall-e-3"
	}

	providerID, modelID, ok := providers.ResolveModelProvider(modelStr, h.DB)
	if !ok {
		providerID = "openai"
		modelID = modelStr
	}

	sel, err := providers.SelectCredentials(h.DB, providerID, nil)
	if err != nil || sel == nil {
		h.JSONError(w, http.StatusNotFound, fmt.Sprintf("no active credentials for provider: %s", providerID))
		return
	}

	body["model"] = modelID
	data, _ := json.Marshal(body)

	baseURL := sel.Credentials.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, baseURL+"/images/generations", bytes.NewReader(data))
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to build request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if sel.Credentials.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+sel.Credentials.APIKey)
	}

	client := &http.Client{Timeout: providers.DefaultClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, "upstream error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) HandleAudioSpeech(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	inputStr, _ := body["input"].(string)
	if strings.TrimSpace(inputStr) == "" {
		h.JSONError(w, http.StatusBadRequest, "Missing required field: input")
		return
	}

	modelStr, _ := body["model"].(string)
	voiceStr, _ := body["voice"].(string)
	targetVoice := voiceStr
	if targetVoice == "" {
		targetVoice = modelStr
	}

	providerID, modelID, hasProvider := providers.ResolveModelProvider(modelStr, h.DB)
	isEdgeTTS := providerID == "edge-tts" ||
		providerID == "local-device" ||
		strings.Contains(strings.ToLower(modelStr), "neural") ||
		strings.Contains(strings.ToLower(voiceStr), "neural") ||
		!hasProvider

	if isEdgeTTS {
		if targetVoice == "" || targetVoice == "edge-tts" || targetVoice == "tts-1" {
			targetVoice = "id-ID-GadisNeural"
		}
		audioBytes, err := providers.SynthesizeEdgeTTS(inputStr, targetVoice)
		if err != nil {
			h.JSONError(w, http.StatusBadGateway, "Edge TTS synthesis failed: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(audioBytes)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(audioBytes)
		return
	}

	sel, err := providers.SelectCredentials(h.DB, providerID, nil)
	if err != nil || sel == nil {
		// Fallback to Edge TTS
		audioBytes, err := providers.SynthesizeEdgeTTS(inputStr, "id-ID-GadisNeural")
		if err != nil {
			h.JSONError(w, http.StatusNotFound, fmt.Sprintf("no active credentials for provider: %s", providerID))
			return
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(audioBytes)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(audioBytes)
		return
	}

	body["model"] = modelID
	data, _ := json.Marshal(body)

	baseURL := sel.Credentials.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, baseURL+"/audio/speech", bytes.NewReader(data))
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to build request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if sel.Credentials.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+sel.Credentials.APIKey)
	}

	client := &http.Client{Timeout: providers.DefaultClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, "upstream error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) HandleAudioTranscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	active := true
	conns, _ := repos.ListConnections(h.DB, repos.ConnectionFilter{IsActive: &active})
	var selConn *repos.Connection
	for _, c := range conns {
		if c.Provider == "groq" || c.Provider == "bynara" || c.Provider == "openai" {
			selConn = &c
			break
		}
	}
	if selConn == nil && len(conns) > 0 {
		selConn = &conns[0]
	}

	if selConn == nil {
		h.JSONError(w, http.StatusNotFound, "no active credentials for speech-to-text")
		return
	}

	baseURL, _ := selConn.Data["baseUrl"].(string)
	apiKey := selConn.APIKey
	if apiKey == "" {
		apiKey, _ = selConn.Data["apiKey"].(string)
	}
	if baseURL == "" {
		if selConn.Provider == "groq" {
			baseURL = "https://api.groq.com/openai/v1"
		} else if selConn.Provider == "bynara" {
			baseURL = "https://router.bynara.id/v1"
		} else {
			baseURL = "https://api.openai.com/v1"
		}
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, baseURL+"/audio/transcriptions", r.Body)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to build request")
		return
	}
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: providers.DefaultClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, "upstream error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	query, _ := body["query"].(string)
	if query == "" {
		h.JSONError(w, http.StatusBadRequest, "Missing required field: query")
		return
	}

	// OpenAI-compatible search response
	results := []map[string]any{
		{
			"title":   "Search result for " + query,
			"url":     "https://9router.com",
			"content": fmt.Sprintf("Simulated 9Router search result for query '%s'", query),
		},
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"query":   query,
		"results": results,
	})
}

func (h *Handler) HandleWebFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		h.JSONError(w, http.StatusBadRequest, "Missing or invalid 'url' field")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, body.URL, nil)
	if err != nil {
		h.JSONError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) 9Router/0.5")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, "fetch failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	contentBytes, _ := io.ReadAll(resp.Body)
	h.JSON(w, http.StatusOK, map[string]any{
		"url":         body.URL,
		"status":      resp.StatusCode,
		"content":     string(contentBytes),
		"contentType": resp.Header.Get("Content-Type"),
	})
}

func (h *Handler) HandleVideoGenerations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	sel, err := providers.SelectCredentials(h.DB, "xai", nil)
	if err != nil || sel == nil {
		h.JSONError(w, http.StatusNotFound, "no active credentials for xai provider")
		return
	}

	baseURL := sel.Credentials.BaseURL
	if baseURL == "" {
		baseURL = "https://api.x.ai/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, baseURL+"/videos/generations", bytes.NewReader(data))
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to build request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if sel.Credentials.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+sel.Credentials.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, "upstream error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) HandleVideoStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/videos/")
	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "Missing video request id")
		return
	}

	sel, err := providers.SelectCredentials(h.DB, "xai", nil)
	if err != nil || sel == nil {
		h.JSONError(w, http.StatusNotFound, "no active credentials for xai provider")
		return
	}

	baseURL := sel.Credentials.BaseURL
	if baseURL == "" {
		baseURL = "https://api.x.ai/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, baseURL+"/videos/"+id, nil)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to build request")
		return
	}
	if sel.Credentials.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+sel.Credentials.APIKey)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, "upstream error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
