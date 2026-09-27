package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProxyPool struct {
	ID                  string         `json:"id"`
	IsActive            bool           `json:"isActive"`
	TestStatus          string         `json:"testStatus,omitempty"`
	Data                map[string]any `json:"-"`
	CreatedAt           string         `json:"createdAt"`
	UpdatedAt           string         `json:"updatedAt"`

	ProxyURL            string `json:"proxyUrl,omitempty"`
	Name                string `json:"name,omitempty"`
	Type                string `json:"type,omitempty"`
	NoProxy             string `json:"noProxy,omitempty"`
	StrictProxy         bool   `json:"strictProxy"`
	LastTestedAt        string `json:"lastTestedAt,omitempty"`
	LastError           string `json:"lastError,omitempty"`
	VercelURL           string `json:"vercelRelayUrl,omitempty"`
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
	isActive := p.IsActive
	if v, ok := updates["isActive"].(bool); ok {
		isActive = v
	} else if v, ok := updates["isActive"].(int); ok {
		isActive = v != 0
	} else if v, ok := updates["isActive"].(float64); ok {
		isActive = v != 0
	}
	testStatus := p.TestStatus
	if v, ok := updates["testStatus"].(string); ok {
		testStatus = v
	}

	updated := dataToPool(id, isActive, testStatus, data, p.CreatedAt)
	updated.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	b, _ := json.Marshal(poolToData(updated))
	activeInt := 0
	if updated.IsActive {
		activeInt = 1
	}
	_, err = db.Exec(
		`UPDATE proxyPools SET isActive=?, testStatus=?, data=?, updatedAt=? WHERE id=?`,
		activeInt, nvl(updated.TestStatus), string(b), updated.UpdatedAt, id,
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

func (p ProxyPool) MarshalJSON() ([]byte, error) {
	m := poolToData(p)
	m["id"] = p.ID
	m["isActive"] = p.IsActive
	if p.TestStatus != "" {
		m["testStatus"] = p.TestStatus
	}
	m["createdAt"] = p.CreatedAt
	m["updatedAt"] = p.UpdatedAt
	return json.Marshal(m)
}

func PoolToData(p ProxyPool) map[string]any {
	return poolToData(p)
}

func poolToData(p ProxyPool) map[string]any {
	m := map[string]any{}
	for k, v := range p.Data {
		m[k] = v
	}
	if p.ProxyURL != "" {
		m["proxyUrl"] = p.ProxyURL
	}
	if p.Name != "" {
		m["name"] = p.Name
	}
	if p.Type != "" {
		m["type"] = p.Type
	}
	if p.NoProxy != "" {
		m["noProxy"] = p.NoProxy
	}
	m["strictProxy"] = p.StrictProxy
	if p.LastTestedAt != "" {
		m["lastTestedAt"] = p.LastTestedAt
	}
	if p.LastError != "" {
		m["lastError"] = p.LastError
	}
	if p.VercelURL != "" {
		m["vercelRelayUrl"] = p.VercelURL
	}
	if p.CloudflareWorkerURL != "" {
		m["cloudflareWorkerUrl"] = p.CloudflareWorkerURL
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
	p.Type, _ = data["type"].(string)
	p.NoProxy, _ = data["noProxy"].(string)
	if sp, ok := data["strictProxy"].(bool); ok {
		p.StrictProxy = sp
	}
	p.LastTestedAt, _ = data["lastTestedAt"].(string)
	p.LastError, _ = data["lastError"].(string)
	p.VercelURL, _ = data["vercelRelayUrl"].(string)
	p.CloudflareWorkerURL, _ = data["cloudflareWorkerUrl"].(string)
	return p
}

