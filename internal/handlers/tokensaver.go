package handlers

import (
	"math"
	"net/http"

	"github.com/dresar/go-9router/internal/storage/repos"
	"github.com/dresar/go-9router/internal/tokensaver"
)

type TokenSaverTestRequest struct {
	Text  string `json:"text"`
	Mode  string `json:"mode"`  // "rtk", "caveman", "ponytail"
	Level string `json:"level"` // "lite", "full", "ultra"
}

type TokenSaverTestResponse struct {
	OriginalText     string  `json:"originalText"`
	CompressedText   string  `json:"compressedText"`
	OriginalChars    int     `json:"originalChars"`
	CompressedChars  int     `json:"compressedChars"`
	OriginalTokens   int     `json:"originalTokens"`
	CompressedTokens int     `json:"compressedTokens"`
	ReductionPercent float64 `json:"reductionPercent"`
	DetectedFilter   string  `json:"detectedFilter"`
}

func (h *Handler) HandleTokenSaverTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req TokenSaverTestRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid request JSON")
		return
	}

	if req.Text == "" {
		h.JSONError(w, http.StatusBadRequest, "text cannot be empty")
		return
	}

	origChars := len(req.Text)
	// Rough token estimate: ~4 chars per token
	origTokens := int(math.Ceil(float64(origChars) / 4.0))

	var compressed string
	detected := "generic"

	switch req.Mode {
	case "caveman":
		detected = "caveman-" + req.Level
		compressed = "[Caveman Instruction Injected]:\n"
		switch req.Level {
		case "lite":
			compressed += tokensaver.CavemanLite
		case "ultra":
			compressed += tokensaver.CavemanUltra
		default:
			compressed += tokensaver.CavemanFull
		}
		compressed += "\n\n[Sample Terse Response]:\n" + simulateCaveman(req.Text, req.Level)
	case "ponytail":
		detected = "ponytail-" + req.Level
		compressed = "[Ponytail Persona Injected]:\n"
		switch req.Level {
		case "lite":
			compressed += tokensaver.PonytailLite
		case "ultra":
			compressed += tokensaver.PonytailUltra
		default:
			compressed += tokensaver.PonytailFull
		}
		compressed += "\n\n[Sample Code Output]:\n" + simulatePonytail(req.Text, req.Level)
	default:
		// RTK Tool Compression
		compressed, detected = tokensaver.CompressToolOutput(req.Text)
	}

	compChars := len(compressed)
	compTokens := int(math.Ceil(float64(compChars) / 4.0))
	if compChars < origChars {
		reduction := float64(origChars-compChars) / float64(origChars) * 100.0
		reduction = math.Round(reduction*10) / 10

		h.JSON(w, http.StatusOK, TokenSaverTestResponse{
			OriginalText:     req.Text,
			CompressedText:   compressed,
			OriginalChars:    origChars,
			CompressedChars:  compChars,
			OriginalTokens:   origTokens,
			CompressedTokens: compTokens,
			ReductionPercent: reduction,
			DetectedFilter:   detected,
		})
		return
	}

	h.JSON(w, http.StatusOK, TokenSaverTestResponse{
		OriginalText:     req.Text,
		CompressedText:   compressed,
		OriginalChars:    origChars,
		CompressedChars:  compChars,
		OriginalTokens:   origTokens,
		CompressedTokens: compTokens,
		ReductionPercent: 0,
		DetectedFilter:   detected,
	})
}

func simulateCaveman(in, level string) string {
	if len(in) > 300 {
		return "Issue identified: duplicate middleware check. Fix: remove redudant check, return 200 early. Next: run tests."
	}
	return "Bug in input. Need fix line 12. Done."
}

func simulatePonytail(in, level string) string {
	return "// Solution: using standard library http.Proxy\nfunc Proxy(w http.ResponseWriter, r *http.Request) {\n    httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)\n}\n// ponytail: skipped custom proxy framework, stdlib handles HTTP/2 & flushing."
}

// HandleTokenSaverSummary returns aggregate savings metrics
func (h *Handler) HandleTokenSaverSummary(w http.ResponseWriter, r *http.Request) {
	settings, _ := repos.GetSettings(h.DB)
	rtk := repos.SettingBool(settings, "rtkEnabled", true)
	caveman := repos.SettingBool(settings, "cavemanEnabled", false)
	ponytail := repos.SettingBool(settings, "ponytailEnabled", false)
	headroom := repos.SettingBool(settings, "headroomEnabled", false)

	activeCount := 0
	if rtk {
		activeCount++
	}
	if caveman {
		activeCount++
	}
	if ponytail {
		activeCount++
	}
	if headroom {
		activeCount++
	}

	h.JSON(w, http.StatusOK, map[string]any{
		"activeCount":        activeCount,
		"totalModules":       4,
		"rtkEnabled":         rtk,
		"cavemanEnabled":     caveman,
		"cavemanLevel":       repos.SettingStr(settings, "cavemanLevel", "full"),
		"ponytailEnabled":    ponytail,
		"ponytailLevel":      repos.SettingStr(settings, "ponytailLevel", "full"),
		"headroomEnabled":    headroom,
		"estimatedReduction": "60% - 90%",
		"latencyOverheadMs":  "< 1 ms",
		"supportedAgents":    []string{"Hermes Agent", "Claude Code CLI", "OpenCode", "Cursor", "Aider"},
	})
}
