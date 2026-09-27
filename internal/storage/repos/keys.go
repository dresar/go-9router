package repos

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	MachineID string `json:"machineId"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
}

func ListAPIKeys(db *sql.DB) ([]APIKey, error) {
	rows, err := db.Query(`SELECT id, key, name, machineId, isActive, createdAt FROM apiKeys ORDER BY createdAt DESC`)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()
	var out []APIKey
	for rows.Next() {
		var k APIKey
		var machineID sql.NullString
		var name sql.NullString
		var isActive int
		if err := rows.Scan(&k.ID, &k.Key, &name, &machineID, &isActive, &k.CreatedAt); err != nil {
			return nil, err
		}
		k.Name = name.String
		k.MachineID = machineID.String
		k.IsActive = isActive == 1
		out = append(out, k)
	}
	return out, rows.Err()
}

func GetAPIKeyByID(db *sql.DB, id string) (*APIKey, error) {
	var k APIKey
	var machineID, name sql.NullString
	var isActive int
	err := db.QueryRow(`SELECT id, key, name, machineId, isActive, createdAt FROM apiKeys WHERE id = ?`, id).
		Scan(&k.ID, &k.Key, &name, &machineID, &isActive, &k.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	k.Name = name.String
	k.MachineID = machineID.String
	k.IsActive = isActive == 1
	return &k, nil
}

func CreateAPIKey(db *sql.DB, name, machineID string) (*APIKey, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	k := APIKey{
		ID:        uuid.NewString(),
		Key:       "9r-" + hex.EncodeToString(raw),
		Name:      name,
		MachineID: machineID,
		IsActive:  true,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	_, err := db.Exec(
		`INSERT INTO apiKeys(id, key, name, machineId, isActive, createdAt) VALUES(?,?,?,?,1,?)`,
		k.ID, k.Key, nvl(k.Name), nvl(k.MachineID), k.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}
	return &k, nil
}

func DeleteAPIKey(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM apiKeys WHERE id = ?`, id)
	return err
}

func ValidateAPIKey(db *sql.DB, key string) (bool, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return false, nil
	}
	var isActive int
	err := db.QueryRow(`SELECT isActive FROM apiKeys WHERE key = ?`, key).Scan(&isActive)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return isActive == 1, nil
}
