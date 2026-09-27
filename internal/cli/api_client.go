package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

func NewAPIClient(host string, port int) *APIClient {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	token := getLocalCliToken()
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Token: token,
	}
}

func getLocalCliToken() string {
	dataDir := GetAppDataDir()
	machineIDPath := filepath.Join(dataDir, "machine-id")
	machineIDBytes, err := os.ReadFile(machineIDPath)
	if err != nil {
		return ""
	}
	machineID := strings.TrimSpace(string(machineIDBytes))

	secretPath := filepath.Join(dataDir, "auth", "cli-secret")
	secretBytes, err := os.ReadFile(secretPath)
	if err != nil {
		return ""
	}
	secret := strings.TrimSpace(string(secretBytes))

	h := sha256.New()
	h.Write([]byte(machineID + "9r-cli-auth" + secret))
	digest := hex.EncodeToString(h.Sum(nil))
	if len(digest) > 16 {
		return digest[:16]
	}
	return digest
}

func (c *APIClient) GetHealth() (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/health", nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("x-9r-cli-token", c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) GetVersion() (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/version", nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("x-9r-cli-token", c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) Shutdown() error {
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/shutdown", nil)
	if err != nil {
		return err
	}
	if c.Token != "" {
		req.Header.Set("x-9r-cli-token", c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}
