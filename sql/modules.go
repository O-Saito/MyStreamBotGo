package sql

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func OpenModuleDB(moduleName string) (*sql.DB, error) {
	if moduleName == "" {
		return nil, fmt.Errorf("OpenModuleDB(%q): invalid module name", moduleName)
	}
	dir := filepath.Join("./db", "modules")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, fmt.Sprintf("%s.db", moduleName))

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`PRAGMA journal_mode = WAL;`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
