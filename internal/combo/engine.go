package combo

import (
	"strings"
	"sync"
)

type RotationState struct {
	Index               int
	ConsecutiveUseCount int
}

type Engine struct {
	mu            sync.RWMutex
	rotationState map[string]*RotationState
}

var DefaultEngine = NewEngine()

func NewEngine() *Engine {
	return &Engine{
		rotationState: make(map[string]*RotationState),
	}
}

// GetRotatedModels applies round-robin or fallback rotation strategy to candidate models.
func (e *Engine) GetRotatedModels(comboName string, models []string, strategy string, stickyLimit int) []string {
	if len(models) <= 1 || strings.ToLower(strategy) != "round-robin" {
		return cloneSlice(models)
	}

	if stickyLimit <= 0 {
		stickyLimit = 1
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	state, exists := e.rotationState[comboName]
	if !exists {
		state = &RotationState{Index: 0, ConsecutiveUseCount: 0}
		e.rotationState[comboName] = state
	}

	currentIndex := state.Index % len(models)
	rotated := make([]string, len(models))
	for i := 0; i < len(models); i++ {
		rotated[i] = models[(currentIndex+i)%len(models)]
	}

	state.ConsecutiveUseCount++
	if state.ConsecutiveUseCount >= stickyLimit {
		state.Index = (currentIndex + 1) % len(models)
		state.ConsecutiveUseCount = 0
	}

	return rotated
}

// ResetRotation clears the in-memory rotation state for a combo or all combos.
func (e *Engine) ResetRotation(comboName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if comboName != "" {
		delete(e.rotationState, comboName)
	} else {
		e.rotationState = make(map[string]*RotationState)
	}
}

// DetectRequiredCapabilities inspects the chat request body for multimodal/vision requirements.
func DetectRequiredCapabilities(body map[string]any) bool {
	if body == nil {
		return false
	}

	// 1. Check messages array (OpenAI / Claude / Ollama / Hermes)
	if msgs, ok := body["messages"].([]any); ok && len(msgs) > 0 {
		// Only inspect the last user turn (matches 9router behavior)
		for i := len(msgs) - 1; i >= 0; i-- {
			m, ok := msgs[i].(map[string]any)
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			if role == "assistant" || role == "model" {
				break
			}
			if checkMessageForVision(m) {
				return true
			}
		}
	}

	// 2. Check Gemini contents
	if contents, ok := body["contents"].([]any); ok {
		for _, c := range contents {
			if cm, ok := c.(map[string]any); ok {
				if parts, ok := cm["parts"].([]any); ok {
					for _, p := range parts {
						if pm, ok := p.(map[string]any); ok {
							if pm["inlineData"] != nil || pm["inline_data"] != nil || pm["fileData"] != nil {
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}

func checkMessageForVision(m map[string]any) bool {
	if m["image_url"] != nil || m["image"] != nil {
		return true
	}
	if imgs, ok := m["images"].([]any); ok && len(imgs) > 0 {
		return true
	}
	if atts, ok := m["attachments"].([]any); ok && len(atts) > 0 {
		return true
	}

	// Check content
	switch c := m["content"].(type) {
	case string:
		if strings.Contains(c, "data:image/") {
			return true
		}
	case []any:
		for _, block := range c {
			if bm, ok := block.(map[string]any); ok {
				bType, _ := bm["type"].(string)
				if bType == "image_url" || bType == "image" || bType == "input_image" {
					return true
				}
				if src, ok := bm["source"].(map[string]any); ok {
					mediaType, _ := src["media_type"].(string)
					if strings.HasPrefix(mediaType, "image/") {
						return true
					}
				}
			}
		}
	}

	return false
}

// IsVisionCapable checks if a model identifier typically supports multimodal/vision.
func IsVisionCapable(modelStr string) bool {
	lower := strings.ToLower(modelStr)
	indicators := []string{
		"gemini", "vision", "gpt-4o", "flash", "vl",
		"claude-3", "claude-sonnet", "claude-opus",
		"qwen-vl", "llava", "pixtral", "omni",
	}
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}
	return false
}

// ReorderByCapabilities floats models satisfying vision/multimodal requirements to the front.
func ReorderByCapabilities(models []string, hasVision bool) []string {
	if !hasVision || len(models) <= 1 {
		return cloneSlice(models)
	}

	var visionModels []string
	var otherModels []string

	for _, m := range models {
		if IsVisionCapable(m) {
			visionModels = append(visionModels, m)
		} else {
			otherModels = append(otherModels, m)
		}
	}

	if len(visionModels) == 0 {
		return cloneSlice(models)
	}

	return append(visionModels, otherModels...)
}

func cloneSlice(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

type PresetItem struct {
	Name   string   `json:"name"`
	Models []string `json:"models"`
	Exists bool     `json:"exists"`
}

// BuildPresets returns standard combo presets for Claude or Cursor.
func BuildPresets(source string, existingNames []string) []PresetItem {
	existingMap := make(map[string]bool, len(existingNames))
	for _, n := range existingNames {
		existingMap[n] = true
	}

	var rawItems []PresetItem
	switch strings.ToLower(source) {
	case "cursor":
		rawItems = []PresetItem{
			{Name: "claude-3.5-sonnet", Models: []string{"cu/claude-3.5-sonnet"}},
			{Name: "claude-3-5-sonnet", Models: []string{"cu/claude-3.5-sonnet"}},
			{Name: "gpt-4o", Models: []string{"cu/gpt-4o"}},
			{Name: "cursor-small", Models: []string{"cu/cursor-small"}},
			{Name: "gemini-2.5-flash", Models: []string{"cu/gemini-2.5-flash"}},
		}
	default: // "claude"
		rawItems = []PresetItem{
			{Name: "default", Models: []string{"cc/claude-sonnet-5"}},
			{Name: "opusplan", Models: []string{"cc/claude-opus-5"}},
			{Name: "claude-3-5-sonnet-20241022", Models: []string{"cc/claude-3-5-sonnet-20241022"}},
			{Name: "claude-3-5-haiku-20241022", Models: []string{"cc/claude-3-5-haiku-20241022"}},
			{Name: "claude-3-opus-20240229", Models: []string{"cc/claude-3-opus-20240229"}},
			{Name: "claude-sonnet-4", Models: []string{"cc/claude-sonnet-4"}},
		}
	}

	for i := range rawItems {
		rawItems[i].Exists = existingMap[rawItems[i].Name]
	}

	return rawItems
}
