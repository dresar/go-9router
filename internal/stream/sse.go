package stream

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ProxySSE(ctx context.Context, w http.ResponseWriter, upstream *http.Response) error {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	reader := bufio.NewReaderSize(upstream.Body, 4096)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line, err := reader.ReadString('\n')
		if line != "" {
			w.Write([]byte(line))
			if strings.HasSuffix(line, "\n") {
				flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func ProxyJSON(upstream *http.Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(upstream.StatusCode)
	_, err := io.Copy(w, upstream.Body)
	return err
}

func IsSSERequest(body map[string]any) bool {
	if stream, ok := body["stream"].(bool); ok {
		return stream
	}
	return false
}
