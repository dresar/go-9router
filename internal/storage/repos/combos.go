package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Combo struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Kind      string   `json:"kind,omitempty"`
	Models    []string `json:"models"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

func ListCombos(db *sql.DB) ([]Combo, error) {
	rows, err := db.Query(`SELECT id, name, kind, models, createdAt, updatedAt FROM combos ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list combos: %w", err)
	}
	defer rows.Close()
	var out []Combo
	for rows.Next() {
		var c Combo
		var kind sql.NullString
		var modelsStr string
		if err := rows.Scan(&c.ID, &c.Name, &kind, &modelsStr, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Kind = kind.String
		_ = json.Unmarshal([]byte(modelsStr), &c.Models)
		out = append(out, c)
	}
	return out, rows.Err()
}

func GetComboByID(db *sql.DB, id string) (*Combo, error) {
	return getCombo(db, `SELECT id, name, kind, models, createdAt, updatedAt FROM combos WHERE id = ?`, id)
}

func GetComboByName(db *sql.DB, name string) (*Combo, error) {
	return getCombo(db, `SELECT id, name, kind, models, createdAt, updatedAt FROM combos WHERE name = ?`, name)
}

func getCombo(db *sql.DB, q, arg string) (*Combo, error) {
	var c Combo
	var kind sql.NullString
	var modelsStr string
	err := db.QueryRow(q, arg).Scan(&c.ID, &c.Name, &kind, &modelsStr, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Kind = kind.String
	_ = json.Unmarshal([]byte(modelsStr), &c.Models)
	return &c, nil
}

func CreateCombo(db *sql.DB, c Combo) (*Combo, error) {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	c.CreatedAt = now
	c.UpdatedAt = now
	mb, _ := json.Marshal(c.Models)
	_, err := db.Exec(
		`INSERT INTO combos(id, name, kind, models, createdAt, updatedAt) VALUES(?,?,?,?,?,?)`,
		c.ID, c.Name, nvl(c.Kind), string(mb), c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create combo: %w", err)
	}
	return &c, nil
}

func UpdateCombo(db *sql.DB, id string, updates map[string]any) (*Combo, error) {
	c, err := GetComboByID(db, id)
	if err != nil || c == nil {
		return nil, err
	}
	if v, ok := updates["name"].(string); ok {
		c.Name = v
	}
	if v, ok := updates["kind"].(string); ok {
		c.Kind = v
	}
	if v, ok := updates["models"].([]string); ok {
		c.Models = v
	}
	c.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	mb, _ := json.Marshal(c.Models)
	_, err = db.Exec(
		`UPDATE combos SET name=?, kind=?, models=?, updatedAt=? WHERE id=?`,
		c.Name, nvl(c.Kind), string(mb), c.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update combo: %w", err)
	}
	return c, nil
}

func DeleteCombo(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM combos WHERE id = ?`, id)
	return err
}
