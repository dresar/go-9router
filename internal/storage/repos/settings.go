package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Settings map[string]any

var (
	settingsCacheMu sync.RWMutex
	cachedSettings  Settings
	settingsExpires time.Time
)

func GetSettings(db *sql.DB) (Settings, error) {
	settingsCacheMu.RLock()
	if cachedSettings != nil && time.Now().Before(settingsExpires) {
		res := make(Settings, len(cachedSettings))
		for k, v := range cachedSettings {
			res[k] = v
		}
		settingsCacheMu.RUnlock()
		return res, nil
	}
	settingsCacheMu.RUnlock()

	var data string
	err := db.QueryRow(`SELECT data FROM settings WHERE id = 1`).Scan(&data)
	if err == sql.ErrNoRows {
		return Settings{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	var s Settings
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return Settings{}, nil
	}

	settingsCacheMu.Lock()
	cachedSettings = s
	settingsExpires = time.Now().Add(3 * time.Second)
	settingsCacheMu.Unlock()

	res := make(Settings, len(s))
	for k, v := range s {
		res[k] = v
	}
	return res, nil
}

func UpdateSettings(db *sql.DB, updates map[string]any) (Settings, error) {
	settingsCacheMu.Lock()
	cachedSettings = nil
	settingsExpires = time.Time{}
	settingsCacheMu.Unlock()

	current, err := GetSettings(db)
	if err != nil {
		return nil, err
	}
	for k, v := range updates {
		if v == nil {
			delete(current, k)
		} else {
			current[k] = v
		}
	}
	data, err := json.Marshal(current)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = db.Exec(
		`INSERT INTO settings(id, data) VALUES(1, ?) ON CONFLICT(id) DO UPDATE SET data = excluded.data`,
		string(data),
	)
	_ = now
	if err != nil {
		return nil, fmt.Errorf("update settings: %w", err)
	}
	return current, nil
}

func SettingBool(s Settings, key string, def bool) bool {
	v, ok := s[key]
	if !ok {
		return def
	}
	switch b := v.(type) {
	case bool:
		return b
	case float64:
		return b != 0
	}
	return def
}

func SettingStr(s Settings, key, def string) string {
	v, ok := s[key]
	if !ok {
		return def
	}
	if str, ok := v.(string); ok {
		return str
	}
	return def
}

func SettingStringSlice(s Settings, key string) []string {
	v, ok := s[key]
	if !ok {
		return nil
	}
	switch sl := v.(type) {
	case []string:
		return sl
	case []any:
		var out []string
		for _, item := range sl {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	}
	return nil
}
