package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProviderNode struct {
	ID        string         `json:"id"`
	Type      string         `json:"type,omitempty"`
	Name      string         `json:"name,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`

	BaseURL string `json:"baseUrl,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
	APIType string `json:"apiType,omitempty"`
}

func ListNodes(db *sql.DB) ([]ProviderNode, error) {
	rows, err := db.Query(`SELECT id, type, name, data, createdAt, updatedAt FROM providerNodes ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProviderNode
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func GetNodeByID(db *sql.DB, id string) (*ProviderNode, error) {
	row := db.QueryRow(`SELECT id, type, name, data, createdAt, updatedAt FROM providerNodes WHERE id = ?`, id)
	n, err := scanNode(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func CreateNode(db *sql.DB, n ProviderNode) (*ProviderNode, error) {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	n.CreatedAt = now
	n.UpdatedAt = now
	data := nodeToData(n)
	b, _ := json.Marshal(data)
	_, err := db.Exec(
		`INSERT INTO providerNodes(id, type, name, data, createdAt, updatedAt) VALUES(?,?,?,?,?,?)`,
		n.ID, nvl(n.Type), nvl(n.Name), string(b), n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create node: %w", err)
	}
	return &n, nil
}

func UpdateNode(db *sql.DB, id string, updates map[string]any) (*ProviderNode, error) {
	n, err := GetNodeByID(db, id)
	if err != nil || n == nil {
		return nil, err
	}
	data := nodeToData(*n)
	for k, v := range updates {
		if v == nil {
			delete(data, k)
		} else {
			data[k] = v
		}
	}
	updated := dataToNode(id, n.Type, n.Name, data, n.CreatedAt)
	updated.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	b, _ := json.Marshal(nodeToData(updated))
	_, err = db.Exec(
		`UPDATE providerNodes SET type=?, name=?, data=?, updatedAt=? WHERE id=?`,
		nvl(updated.Type), nvl(updated.Name), string(b), updated.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update node: %w", err)
	}
	return &updated, nil
}

func DeleteNode(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM providerNodes WHERE id = ?`, id)
	return err
}

func scanNode(row interface{ Scan(...any) error }) (ProviderNode, error) {
	var id, createdAt, updatedAt string
	var typ, name sql.NullString
	var dataStr string
	if err := row.Scan(&id, &typ, &name, &dataStr, &createdAt, &updatedAt); err != nil {
		return ProviderNode{}, err
	}
	var data map[string]any
	_ = json.Unmarshal([]byte(dataStr), &data)
	n := dataToNode(id, typ.String, name.String, data, createdAt)
	n.UpdatedAt = updatedAt
	return n, nil
}

func nodeToData(n ProviderNode) map[string]any {
	m := map[string]any{}
	if n.BaseURL != "" {
		m["baseUrl"] = n.BaseURL
	}
	if n.Prefix != "" {
		m["prefix"] = n.Prefix
	}
	if n.APIType != "" {
		m["apiType"] = n.APIType
	}
	for k, v := range n.Data {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
	return m
}

func dataToNode(id, typ, name string, data map[string]any, createdAt string) ProviderNode {
	n := ProviderNode{ID: id, Type: typ, Name: name, Data: data, CreatedAt: createdAt}
	if data == nil {
		return n
	}
	n.BaseURL, _ = data["baseUrl"].(string)
	n.Prefix, _ = data["prefix"].(string)
	n.APIType, _ = data["apiType"].(string)
	return n
}
