package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func KVGet(db *sql.DB, scope, key string) (any, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM kv WHERE scope = ? AND key = ?`, scope, key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var v any
	_ = json.Unmarshal([]byte(value), &v)
	return v, nil
}

func KVSet(db *sql.DB, scope, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT OR REPLACE INTO kv(scope, key, value) VALUES(?,?,?)`, scope, key, string(b))
	return err
}

func KVDelete(db *sql.DB, scope, key string) error {
	_, err := db.Exec(`DELETE FROM kv WHERE scope = ? AND key = ?`, scope, key)
	return err
}

func KVList(db *sql.DB, scope string) (map[string]any, error) {
	rows, err := db.Query(`SELECT key, value FROM kv WHERE scope = ?`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]any{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		var val any
		_ = json.Unmarshal([]byte(v), &val)
		out[k] = val
	}
	return out, rows.Err()
}

func KVListSlice(db *sql.DB, scope string) ([]any, error) {
	rows, err := db.Query(`SELECT value FROM kv WHERE scope = ? ORDER BY key`, scope)
	if err != nil {
		return nil, fmt.Errorf("kv list slice: %w", err)
	}
	defer rows.Close()
	var out []any
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		var val any
		_ = json.Unmarshal([]byte(v), &val)
		out = append(out, val)
	}
	return out, rows.Err()
}

func KVDeleteScope(db *sql.DB, scope string) error {
	_, err := db.Exec(`DELETE FROM kv WHERE scope = ?`, scope)
	return err
}
