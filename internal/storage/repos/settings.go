package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Settings map[string]any

func GetSettings(db *sql.DB) (Settings, error) {
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
	return s, nil
}

func UpdateSettings(db *sql.DB, updates map[string]any) (Settings, error) {
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
