package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	once     sync.Once
	globalDB *sql.DB
	initErr  error
)

func Open(dataDir string) (*sql.DB, error) {
	once.Do(func() {
		globalDB, initErr = openDB(dataDir)
	})
	if initErr != nil {
		return nil, initErr
	}
	return globalDB, nil
}

func openDB(dataDir string) (*sql.DB, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(dataDir, "db.sqlite")
	db, err := sql.Open("sqlite", path+"?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=ON&_temp_store=MEMORY")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec(`PRAGMA mmap_size = 30000000; PRAGMA cache_size = -64000;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite pragma: %w", err)
	}
	return db, nil
}

func Close() {
	if globalDB != nil {
		globalDB.Close()
	}
}
