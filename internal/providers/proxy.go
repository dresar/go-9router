package providers

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dresar/go-9router/internal/logging"
	"github.com/dresar/go-9router/internal/storage/repos"
)

type ProxyMode string

const (
	ProxyModeDirect ProxyMode = "direct"
	ProxyModeRelay  ProxyMode = "relay"
	ProxyModeHTTP   ProxyMode = "http"
)

type ResolvedProxy struct {
	Mode        ProxyMode
	RelayURL    string
	ProxyURL    string
	NoProxy     string
	StrictProxy bool
	PoolID      string
	PoolName    string
	PoolType    string
}

var (
	proxyClientsMu sync.RWMutex
	proxyClients   = make(map[string]*http.Client)
)

func getProxyHTTPClient(proxyURL string) (*http.Client, error) {
	proxyClientsMu.RLock()
	c, ok := proxyClients[proxyURL]
	proxyClientsMu.RUnlock()
	if ok {
		return c, nil
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL %q: %w", proxyURL, err)
	}

	proxyClientsMu.Lock()
	defer proxyClientsMu.Unlock()
	if c, ok = proxyClients[proxyURL]; ok {
		return c, nil
	}

	transport := &http.Transport{
		Proxy:               http.ProxyURL(parsed),
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Timeout:   0,
		Transport: transport,
	}
	proxyClients[proxyURL] = client
	return client, nil
}

func shouldBypassNoProxy(targetHost, noProxy string) bool {
	if noProxy == "" {
		return false
	}
	targetHost = strings.ToLower(strings.Split(targetHost, ":")[0])
	for _, pattern := range strings.Split(noProxy, ",") {
		p := strings.TrimSpace(strings.ToLower(pattern))
		if p == "" {
			continue
		}
		if p == "*" {
			return true
		}
		if strings.HasPrefix(p, ".") {
			if strings.HasSuffix(targetHost, p) || targetHost == strings.TrimPrefix(p, ".") {
				return true
			}
		} else if targetHost == p || strings.HasSuffix(targetHost, "."+p) {
			return true
		}
	}
	return false
}

func ResolveProxy(db *sql.DB, creds *Credentials, targetURL string) (*ResolvedProxy, error) {
	poolID := ""
	if creds != nil {
		poolID = creds.ProxyPoolID
		if poolID == "" && creds.ProviderSpecificData != nil {
			poolID, _ = creds.ProviderSpecificData["proxyPoolId"].(string)
		}
	}
	if poolID == "__none__" {
		poolID = ""
	}

	if poolID != "" && db != nil {
		pool, err := repos.GetProxyPoolByID(db, poolID)
		if err == nil && pool != nil && pool.IsActive && pool.ProxyURL != "" {
			pType := strings.ToLower(pool.Type)
			if pType == "vercel" || pType == "cloudflare" || pType == "deno" {
				return &ResolvedProxy{
					Mode:        ProxyModeRelay,
					RelayURL:    pool.ProxyURL,
					StrictProxy: pool.StrictProxy,
					PoolID:      pool.ID,
					PoolName:    pool.Name,
					PoolType:    pool.Type,
				}, nil
			}
			return &ResolvedProxy{
				Mode:        ProxyModeHTTP,
				ProxyURL:    pool.ProxyURL,
				NoProxy:     pool.NoProxy,
				StrictProxy: pool.StrictProxy,
				PoolID:      pool.ID,
				PoolName:    pool.Name,
				PoolType:    pool.Type,
			}, nil
		}
	}

	if creds != nil && creds.ProviderSpecificData != nil {
		if enabled, _ := creds.ProviderSpecificData["connectionProxyEnabled"].(bool); enabled {
			proxyURL, _ := creds.ProviderSpecificData["connectionProxyUrl"].(string)
			noProxy, _ := creds.ProviderSpecificData["connectionNoProxy"].(string)
			if proxyURL != "" {
				return &ResolvedProxy{
					Mode:     ProxyModeHTTP,
					ProxyURL: proxyURL,
					NoProxy:  noProxy,
				}, nil
			}
		}
	}

	// MITM hosts bypass: if target is Google cloudcode-pa, direct connections from Indonesia
	// get connection reset by remote host. Automatically route via any active relay pool.
	lowerTarget := strings.ToLower(targetURL)
	if (strings.Contains(lowerTarget, "cloudcode-pa.googleapis.com") || strings.Contains(lowerTarget, "daily-cloudcode-pa.googleapis.com")) && db != nil {
		activePools, err := repos.ListProxyPools(db, true)
		if err == nil {
			for _, p := range activePools {
				pType := strings.ToLower(p.Type)
				if (pType == "vercel" || pType == "cloudflare") && p.ProxyURL != "" && p.TestStatus != "error" {
					return &ResolvedProxy{
						Mode:        ProxyModeRelay,
						RelayURL:    p.ProxyURL,
						StrictProxy: false,
						PoolID:      p.ID,
						PoolName:    p.Name,
						PoolType:    p.Type,
					}, nil
				}
			}
		}
	}

	return &ResolvedProxy{Mode: ProxyModeDirect}, nil
}

func DoUpstreamWithProxy(ctx context.Context, req *http.Request, db *sql.DB, creds *Credentials) (*UpstreamResult, error) {
	req = req.WithContext(ctx)

	targetURL := req.URL.String()
	resolved, err := ResolveProxy(db, creds, targetURL)
	if err != nil {
		return nil, fmt.Errorf("resolve proxy: %w", err)
	}

	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	origURL := *req.URL
	origHost := req.Host

	switch resolved.Mode {
	case ProxyModeRelay:
		targetBase := fmt.Sprintf("%s://%s", origURL.Scheme, origURL.Host)
		targetPath := origURL.Path
		if origURL.RawQuery != "" {
			targetPath += "?" + origURL.RawQuery
		}

		parsedRelay, err := url.Parse(resolved.RelayURL)
		if err != nil {
			if resolved.StrictProxy {
				return nil, fmt.Errorf("invalid relay url: %w", err)
			}
			return DoUpstream(ctx, req)
		}

		req.Header.Set("x-relay-target", targetBase)
		req.Header.Set("x-relay-path", targetPath)
		req.URL.Scheme = parsedRelay.Scheme
		req.URL.Host = parsedRelay.Host
		req.URL.Path = parsedRelay.Path
		req.URL.RawQuery = parsedRelay.RawQuery
		req.Host = parsedRelay.Host

		logging.Info("PROXY", fmt.Sprintf("▶ RELAY [%s/%s] %s -> %s%s", resolved.PoolType, resolved.PoolName, resolved.RelayURL, targetBase, targetPath))

		if len(bodyBytes) > 0 {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		res, err := DoUpstream(ctx, req)
		if (err != nil || (res != nil && (res.Status == 404 || res.Status == 500 || res.Status == 502 || res.Status == 503))) && !resolved.StrictProxy && db != nil {
			// Failover: if this relay failed with network error or dead deployment, try another active relay
			logging.Warn("PROXY", fmt.Sprintf("Relay %s failed, attempting failover...", resolved.RelayURL))
			activePools, _ := repos.ListProxyPools(db, true)
			for _, alt := range activePools {
				altType := strings.ToLower(alt.Type)
				if alt.ID != resolved.PoolID && (altType == "vercel" || altType == "cloudflare" || altType == "deno") && alt.ProxyURL != "" && alt.TestStatus != "error" {
					if altRelay, err2 := url.Parse(alt.ProxyURL); err2 == nil {
						req.URL.Scheme = altRelay.Scheme
						req.URL.Host = altRelay.Host
						req.URL.Path = altRelay.Path
						req.URL.RawQuery = altRelay.RawQuery
						req.Host = altRelay.Host
						if len(bodyBytes) > 0 {
							req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
						}
						logging.Info("PROXY", fmt.Sprintf("▶ FAILOVER RELAY [%s/%s] %s -> %s%s", alt.Type, alt.Name, alt.ProxyURL, targetBase, targetPath))
						altRes, altErr := DoUpstream(ctx, req)
						if altErr == nil && altRes != nil && altRes.Status < 500 && altRes.Status != 404 {
							return altRes, nil
						}
					}
				}
			}

			// All relays failed: fall back to direct request if strictProxy is not set
			logging.Warn("PROXY", fmt.Sprintf("All relays failed, falling back to direct connection for %s", targetBase))
			req.URL = &origURL
			req.Host = origHost
			req.Header.Del("x-relay-target")
			req.Header.Del("x-relay-path")
			if len(bodyBytes) > 0 {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			return DoUpstream(ctx, req)
		}
		return res, err

	case ProxyModeHTTP:
		if shouldBypassNoProxy(req.URL.Host, resolved.NoProxy) {
			logging.Info("PROXY", fmt.Sprintf("Bypassing proxy for %s via no_proxy rule", req.URL.Host))
			return DoUpstream(ctx, req)
		}

		client, err := getProxyHTTPClient(resolved.ProxyURL)
		if err != nil {
			if resolved.StrictProxy {
				return nil, fmt.Errorf("get proxy client: %w", err)
			}
			return DoUpstream(ctx, req)
		}

		logging.Info("PROXY", fmt.Sprintf("▶ PROXY [%s] %s -> %s", resolved.PoolName, resolved.ProxyURL, req.URL.Host))
		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return &UpstreamResult{Error: "request cancelled", Status: 499, Success: false}, nil
			}
			if resolved.StrictProxy {
				return nil, fmt.Errorf("proxy upstream error (strictProxy=true): %w", err)
			}
			logging.Warn("PROXY", fmt.Sprintf("Proxy failed: %v, falling back to direct", err))
			return DoUpstream(ctx, req)
		}

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(body))
			return &UpstreamResult{
				Response: resp,
				Error:    string(body),
				Status:   resp.StatusCode,
				Success:  false,
			}, nil
		}
		return &UpstreamResult{Response: resp, Status: resp.StatusCode, Success: true}, nil

	default:
		return DoUpstream(ctx, req)
	}
}

func TestProxyPool(p *repos.ProxyPool) (ok bool, status int, elapsedMs int64, errStr string) {
	if p == nil || p.ProxyURL == "" {
		return false, 400, 0, "empty proxy URL"
	}

	pType := strings.ToLower(p.Type)
	startedAt := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if pType == "vercel" || pType == "cloudflare" || pType == "deno" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.ProxyURL, nil)
		if err != nil {
			return false, 500, time.Since(startedAt).Milliseconds(), err.Error()
		}
		req.Header.Set("x-relay-target", "https://httpbin.org")
		req.Header.Set("x-relay-path", "/get")
		req.Header.Set("User-Agent", "9Router-Relay-Test/2.0")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		elapsed := time.Since(startedAt).Milliseconds()
		if err != nil {
			return false, 500, elapsed, err.Error()
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return true, resp.StatusCode, elapsed, ""
		}
		return false, resp.StatusCode, elapsed, fmt.Sprintf("relay returned status %d", resp.StatusCode)
	}

	parsed, err := url.Parse(p.ProxyURL)
	if err != nil {
		return false, 400, 0, fmt.Sprintf("invalid proxy URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsed),
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://httpbin.org/get", nil)
	if err != nil {
		return false, 500, time.Since(startedAt).Milliseconds(), err.Error()
	}

	resp, err := client.Do(req)
	elapsed := time.Since(startedAt).Milliseconds()
	if err != nil {
		return false, 500, elapsed, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, resp.StatusCode, elapsed, ""
	}
	return false, resp.StatusCode, elapsed, fmt.Sprintf("proxy test failed with status %d", resp.StatusCode)
}
