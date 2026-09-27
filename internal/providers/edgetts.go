package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	bingVoicesURL     = "https://speech.platform.bing.com/consumer/speech/synthesize/readaloud/voices/list?trustedclienttoken=6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	bingTranslatorURL = "https://www.bing.com/translator"
	bingTtsURL        = "https://www.bing.com/tfettts?isVertical=1&&IG=1&IID=translator.5023&SFX=1"
	edgeUA            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
)

var (
	edgeVoicesMu    sync.Mutex
	edgeVoicesCache []map[string]any
	edgeLangCache   []map[string]any
	edgeByLangCache map[string]any
	edgeCacheTime   time.Time

	bingTokenMu   sync.Mutex
	bingTokenKey  string
	bingTokenVal  string
	bingCookie    string
	bingTokenTime time.Time
)

type rawVoice struct {
	Name         string `json:"Name"`
	ShortName    string `json:"ShortName"`
	Gender       string `json:"Gender"`
	Locale       string `json:"Locale"`
	FriendlyName string `json:"FriendlyName"`
}

var langNames = map[string]string{
	"id": "Indonesian",
	"en": "English",
	"ja": "Japanese",
	"zh": "Chinese",
	"es": "Spanish",
	"fr": "French",
	"de": "German",
	"ar": "Arabic",
	"ko": "Korean",
	"ru": "Russian",
	"pt": "Portuguese",
	"it": "Italian",
	"vi": "Vietnamese",
	"th": "Thai",
	"hi": "Hindi",
}

func GetEdgeTtsVoices() ([]map[string]any, []map[string]any, map[string]any, error) {
	edgeVoicesMu.Lock()
	defer edgeVoicesMu.Unlock()

	if len(edgeVoicesCache) > 0 && time.Since(edgeCacheTime) < 24*time.Hour {
		return edgeVoicesCache, edgeLangCache, edgeByLangCache, nil
	}

	req, err := http.NewRequest(http.MethodGet, bingVoicesURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", edgeUA)
		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var raw []rawVoice
			if err := json.NewDecoder(resp.Body).Decode(&raw); err == nil && len(raw) > 0 {
				buildVoiceCaches(raw)
				return edgeVoicesCache, edgeLangCache, edgeByLangCache, nil
			}
		}
	}

	// Fallback to builtin list if network request fails
	buildVoiceCaches(fallbackVoices())
	return edgeVoicesCache, edgeLangCache, edgeByLangCache, nil
}

func buildVoiceCaches(raw []rawVoice) {
	voices := make([]map[string]any, 0, len(raw))
	byLang := make(map[string]any)

	for _, v := range raw {
		parts := strings.Split(v.Locale, "-")
		lang := parts[0]
		country := ""
		if len(parts) > 1 {
			country = parts[1]
		}
		lName := langNames[lang]
		if lName == "" {
			lName = strings.ToUpper(lang)
		}
		cName := country
		if country == "ID" {
			cName = "Indonesia"
		} else if country == "US" {
			cName = "United States"
		} else if country == "GB" {
			cName = "United Kingdom"
		}

		friendly := v.FriendlyName
		if friendly == "" {
			friendly = v.ShortName
		}
		friendly = strings.ReplaceAll(friendly, "Microsoft ", "")
		friendly = strings.ReplaceAll(friendly, " Online (Natural) - ", " (")

		voiceItem := map[string]any{
			"id":          v.ShortName,
			"name":        friendly,
			"locale":      v.Locale,
			"lang":        lang,
			"country":     country,
			"countryName": cName,
			"langName":    lName,
			"gender":      v.Gender,
		}
		voices = append(voices, voiceItem)

		langGroup, exists := byLang[lang].(map[string]any)
		if !exists {
			langGroup = map[string]any{
				"code":   lang,
				"name":   lName,
				"voices": []map[string]any{},
			}
			byLang[lang] = langGroup
		}
		vList := langGroup["voices"].([]map[string]any)
		langGroup["voices"] = append(vList, voiceItem)
	}

	languages := make([]map[string]any, 0, len(byLang))
	for _, v := range byLang {
		languages = append(languages, v.(map[string]any))
	}
	sort.Slice(languages, func(i, j int) bool {
		nameI := languages[i]["name"].(string)
		nameJ := languages[j]["name"].(string)
		return nameI < nameJ
	})

	edgeVoicesCache = voices
	edgeLangCache = languages
	edgeByLangCache = byLang
	edgeCacheTime = time.Now()
}

func fallbackVoices() []rawVoice {
	return []rawVoice{
		{ShortName: "id-ID-GadisNeural", FriendlyName: "Gadis (Indonesian)", Locale: "id-ID", Gender: "Female"},
		{ShortName: "id-ID-ArdiNeural", FriendlyName: "Ardi (Indonesian)", Locale: "id-ID", Gender: "Male"},
		{ShortName: "en-US-JennyNeural", FriendlyName: "Jenny (US English)", Locale: "en-US", Gender: "Female"},
		{ShortName: "en-US-GuyNeural", FriendlyName: "Guy (US English)", Locale: "en-US", Gender: "Male"},
		{ShortName: "en-US-AriaNeural", FriendlyName: "Aria (US English)", Locale: "en-US", Gender: "Female"},
		{ShortName: "en-US-ChristopherNeural", FriendlyName: "Christopher (US English)", Locale: "en-US", Gender: "Male"},
		{ShortName: "en-GB-SoniaNeural", FriendlyName: "Sonia (UK English)", Locale: "en-GB", Gender: "Female"},
		{ShortName: "en-GB-RyanNeural", FriendlyName: "Ryan (UK English)", Locale: "en-GB", Gender: "Male"},
	}
}

func getBingToken() (string, string, string, error) {
	bingTokenMu.Lock()
	defer bingTokenMu.Unlock()

	if bingTokenVal != "" && time.Since(bingTokenTime) < 5*time.Minute {
		return bingTokenKey, bingTokenVal, bingCookie, nil
	}

	req, err := http.NewRequest(http.MethodGet, bingTranslatorURL, nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("User-Agent", edgeUA)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	var cookies []string
	for _, c := range resp.Cookies() {
		cookies = append(cookies, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	cookieStr := strings.Join(cookies, "; ")

	bodyBytes, _ := io.ReadAll(resp.Body)
	html := string(bodyBytes)

	re := regexp.MustCompile(`params_AbusePreventionHelper\s*=\s*\[([^,]+),([^,]+),`)
	m := re.FindStringSubmatch(html)
	if len(m) < 3 {
		return "", "", "", fmt.Errorf("failed to parse bing abuse prevention token")
	}

	key := strings.TrimSpace(m[1])
	val := strings.Trim(strings.TrimSpace(m[2]), `"`)

	bingTokenKey = key
	bingTokenVal = val
	bingCookie = cookieStr
	bingTokenTime = time.Now()
	return key, val, cookieStr, nil
}

func SynthesizeEdgeTTS(text, voice string) ([]byte, error) {
	if voice == "" {
		voice = "id-ID-GadisNeural"
	}
	parts := strings.Split(voice, "-")
	xmlLang := "id-ID"
	if len(parts) >= 2 {
		xmlLang = parts[0] + "-" + parts[1]
	}
	gender := "Female"
	if strings.Contains(strings.ToLower(voice), "male") || strings.Contains(strings.ToLower(voice), "ardi") || strings.Contains(strings.ToLower(voice), "guy") || strings.Contains(strings.ToLower(voice), "christopher") {
		gender = "Male"
	}

	key, token, cookie, err := getBingToken()
	if err != nil {
		return nil, fmt.Errorf("get bing token: %w", err)
	}

	ssml := fmt.Sprintf(`<speak version='1.0' xml:lang='%s'><voice xml:lang='%s' xml:gender='%s' name='%s'><prosody rate='0.00%%'>%s</prosody></voice></speak>`,
		xmlLang, xmlLang, gender, voice, text)

	formData := url.Values{}
	formData.Set("ssml", ssml)
	formData.Set("token", token)
	formData.Set("key", key)

	req, err := http.NewRequest(http.MethodPost, bingTtsURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", edgeUA)
	req.Header.Set("Origin", "https://www.bing.com")
	req.Header.Set("Referer", "https://www.bing.com/translator")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bingTokenMu.Lock()
		bingTokenVal = ""
		bingTokenMu.Unlock()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("edge tts error %d: %s", resp.StatusCode, string(b))
	}

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(audioBytes) < 500 {
		return nil, fmt.Errorf("edge tts returned empty or invalid audio data")
	}
	return audioBytes, nil
}
