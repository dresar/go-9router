package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type XaiVideoOptions struct {
	Prompt      string
	Output      string
	Model       string
	Duration    int
	AspectRatio string
	Resolution  string
	Image       string
	TimeoutSec  int
	Port        int
	Host        string
	APIKey      string
}

func RunXaiVideo(args []string) error {
	fs := flag.NewFlagSet("xai video", flag.ContinueOnError)
	prompt := fs.String("prompt", "", "Video description (required)")
	output := fs.String("output", "video.mp4", "Output MP4 path")
	model := fs.String("model", "xai/grok-imagine-video", "Model")
	duration := fs.Int("duration", 0, "Video duration in seconds")
	aspectRatio := fs.String("aspect-ratio", "", "Aspect ratio e.g. 16:9, 9:16, 1:1")
	resolution := fs.String("resolution", "", "Resolution 480p | 720p | 1080p")
	image := fs.String("image", "", "Image input for image-to-video")
	timeoutSec := fs.Int("timeout", 600, "Max wait in seconds")
	port := fs.Int("port", 20128, "Gateway port")
	host := fs.String("host", "127.0.0.1", "Gateway host")
	apiKey := fs.String("api-key", os.Getenv("NINE_ROUTER_API_KEY"), "9router API key")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*prompt) == "" {
		return errors.New("missing required argument: --prompt")
	}

	opts := XaiVideoOptions{
		Prompt:      *prompt,
		Output:      *output,
		Model:       *model,
		Duration:    *duration,
		AspectRatio: *aspectRatio,
		Resolution:  *resolution,
		Image:       *image,
		TimeoutSec:  *timeoutSec,
		Port:        *port,
		Host:        *host,
		APIKey:      *apiKey,
	}

	return executeXaiVideo(opts)
}

func executeXaiVideo(opts XaiVideoOptions) error {
	baseURL := fmt.Sprintf("http://%s:%d", opts.Host, opts.Port)
	fmt.Printf("🎬 Submitting xAI Grok Imagine video generation...\n")
	fmt.Printf("   Prompt: %s\n", opts.Prompt)
	fmt.Printf("   Model:  %s\n", opts.Model)

	payload := map[string]any{
		"prompt": opts.Prompt,
		"model":  opts.Model,
	}
	if opts.Duration > 0 {
		payload["duration"] = opts.Duration
	}
	if opts.AspectRatio != "" {
		payload["aspect_ratio"] = opts.AspectRatio
	}
	if opts.Resolution != "" {
		payload["resolution"] = opts.Resolution
	}
	if opts.Image != "" {
		payload["image"] = opts.Image
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/videos/generations", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if opts.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+opts.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("submit video generation: %w (is 9router running on %s:%d?)", err, opts.Host, opts.Port)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstream error (HTTP %d): %s", resp.StatusCode, string(respData))
	}

	var submitResult struct {
		RequestID string `json:"request_id"`
		ID        string `json:"id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(respData, &submitResult); err != nil {
		return fmt.Errorf("parse submit response: %w", err)
	}

	reqID := submitResult.RequestID
	if reqID == "" {
		reqID = submitResult.ID
	}
	if reqID == "" {
		return fmt.Errorf("no request_id returned: %s", string(respData))
	}

	fmt.Printf("✅ Job submitted! Request ID: %s\n", reqID)
	fmt.Printf("⏳ Polling status (timeout: %ds)...\n", opts.TimeoutSec)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opts.TimeoutSec)*time.Second)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	pollURL := fmt.Sprintf("%s/v1/videos/%s", baseURL, reqID)

	for {
		select {
		case <-ctx.Done():
			return errors.New("timed out waiting for video completion")
		case <-ticker.C:
			pollReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, pollURL, nil)
			if opts.APIKey != "" {
				pollReq.Header.Set("Authorization", "Bearer "+opts.APIKey)
			}
			pollResp, err := client.Do(pollReq)
			if err != nil {
				fmt.Printf("   ... network retry: %v\n", err)
				continue
			}

			pollBody, _ := io.ReadAll(pollResp.Body)
			pollResp.Body.Close()

			var statusResult struct {
				Status string `json:"status"`
				Video  struct {
					URL string `json:"url"`
				} `json:"video"`
				URL   string `json:"url"`
				Error any    `json:"error"`
			}
			_ = json.Unmarshal(pollBody, &statusResult)

			fmt.Printf("   Status: %s\n", statusResult.Status)

			switch strings.ToLower(statusResult.Status) {
			case "done", "completed":
				videoURL := statusResult.Video.URL
				if videoURL == "" {
					videoURL = statusResult.URL
				}
				if videoURL == "" {
					return fmt.Errorf("video completed but no download URL found: %s", string(pollBody))
				}
				fmt.Printf("📥 Downloading generated video to %s...\n", opts.Output)
				return downloadFile(videoURL, opts.Output)
			case "failed", "error", "expired", "cancelled":
				return fmt.Errorf("video generation failed with status: %s (error: %v)", statusResult.Status, statusResult.Error)
			}
		}
	}
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	tmpFile := destPath + ".tmp"
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil && filepath.Dir(destPath) != "." {
		return err
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	out.Close()

	if err := os.Rename(tmpFile, destPath); err != nil {
		return err
	}

	fmt.Printf("🎉 Video successfully saved to %s\n", destPath)
	return nil
}
