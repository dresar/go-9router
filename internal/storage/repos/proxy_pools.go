package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProxyPool struct {
	ID         string         `json:"id"`
	IsActive   bool           `json:"isActive"`
	TestStatus string         `json:"testStatus,omitempty"`
	Data       map[string]any `json:"-"`
	CreatedAt  string         `json:"createdAt"`
	UpdatedAt  string         `json:"updatedAt"`

	ProxyURL    string `json:"proxyUrl,omitempty"`
	Name        string `json:"name,omitempty"`
	VercelURL   string `json:"vercelRelayUrl,omitempty"`
	CloudflareWorkerURL string `json:"cloudflareWorkerUrl,omitempty"`
}

func ListProxyPools(db *sql.DB, activeOnly bool) ([]ProxyPool, error) {
	q := `SELECT id, isActive, testStatus, data, createdAt, updatedAt FROM proxyPools`
	if activeOnly {
		q += ` WHERE isActive = 1`
	}
	q += ` ORDER BY createdAt ASC`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProxyPool
	for rows.Next() {
		p, err := scanPool(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetProxyPoolByID(db *sql.DB, id string) (*ProxyPool, error) {
	row := db.QueryRow(`SELECT id, isActive, testStatus, data, createdAt, updatedAt FROM proxyPools WHERE id = ?`, id)
	p, err := scanPool(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func CreateProxyPool(db *sql.DB, p ProxyPool) (*ProxyPool, error) {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	p.CreatedAt = now
	p.UpdatedAt = now
	data := poolToData(p)
	b, _ := json.Marshal(data)
	isActive := 1
	if !p.IsActive {
		isActive = 0
	}
	_, err := db.Exec(
		`INSERT INTO proxyPools(id, isActive, testStatus, data, createdAt, updatedAt) VALUES(?,?,?,?,?,?)`,
		p.ID, isActive, nvl(p.TestStatus), string(b), p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create proxy pool: %w", err)
	}
	return &p, nil
}

func UpdateProxyPool(db *sql.DB, id string, updates map[string]any) (*ProxyPool, error) {
	p, err := GetProxyPoolByID(db, id)
	if err != nil || p == nil {
		return nil, err
	}
	data := poolToData(*p)
	for k, v := range updates {
		if v == nil {
			delete(data, k)
		} else {
			data[k] = v
		}
	}
	updated := dataToPool(id, p.IsActive, p.TestStatus, data, p.CreatedAt)
	updated.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	b, _ := json.Marshal(poolToData(updated))
	isActive := 1
	if !updated.IsActive {
		isActive = 0
	}
	_, err = db.Exec(
		`UPDATE proxyPools SET isActive=?, testStatus=?, data=?, updatedAt=? WHERE id=?`,
		isActive, nvl(updated.TestStatus), string(b), updated.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update proxy pool: %w", err)
	}
	return &updated, nil
}

func DeleteProxyPool(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM proxyPools WHERE id = ?`, id)
	return err
}

func scanPool(row interface{ Scan(...any) error }) (ProxyPool, error) {
	var id, createdAt, updatedAt string
	var testStatus sql.NullString
	var isActive int
	var dataStr string
	if err := row.Scan(&id, &isActive, &testStatus, &dataStr, &createdAt, &updatedAt); err != nil {
		return ProxyPool{}, err
	}
	var data map[string]any
	_ = json.Unmarshal([]byte(dataStr), &data)
	p := dataToPool(id, isActive == 1, testStatus.String, data, createdAt)
	p.UpdatedAt = updatedAt
	return p, nil
}

func poolToData(p ProxyPool) map[string]any {
	m := map[string]any{}
	if p.ProxyURL != "" {
		m["proxyUrl"] = p.ProxyURL
	}
	if p.Name != "" {
		m["name"] = p.Name
	}
	if p.VercelURL != "" {
		m["vercelRelayUrl"] = p.VercelURL
	}
	if p.CloudflareWorkerURL != "" {
		m["cloudflareWorkerUrl"] = p.CloudflareWorkerURL
	}
	for k, v := range p.Data {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
	return m
}

func dataToPool(id string, isActive bool, testStatus string, data map[string]any, createdAt string) ProxyPool {
	p := ProxyPool{ID: id, IsActive: isActive, TestStatus: testStatus, Data: data, CreatedAt: createdAt}
	if data == nil {
		return p
	}
	p.ProxyURL, _ = data["proxyUrl"].(string)
	p.Name, _ = data["name"].(string)
	p.VercelURL, _ = data["vercelRelayUrl"].(string)
	p.CloudflareWorkerURL, _ = data["cloudflareWorkerUrl"].(string)
	return p
}
