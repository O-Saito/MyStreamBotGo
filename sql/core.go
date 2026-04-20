package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type CoreDB struct {
	db *sql.DB
}

func NewCoreDB(fileName string) (*CoreDB, error) {
	dir := filepath.Join("./db")
	corePath := filepath.Join(dir, fmt.Sprintf("%s.db", fileName))

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", corePath)
	if err != nil {
		return nil, err
	}

	// WAL melhora segurança e velocidade
	_, _ = db.Exec(`PRAGMA journal_mode = WAL;`)
	_, _ = db.Exec(`PRAGMA busy_timeout = 5000;`)

	core := &CoreDB{db: db}

	if err := core.initSchema(); err != nil {
		return nil, err
	}

	return core, nil
}

func (c *CoreDB) initSchema() error {
	_, err := c.db.Exec(`
        CREATE TABLE IF NOT EXISTS tokens (
            provider TEXT PRIMARY KEY,
			access_token TEXT,
            refresh_token TEXT,
            expires_at INTEGER
        );

        CREATE TABLE IF NOT EXISTS kv (
            key TEXT PRIMARY KEY,
            value TEXT
        );
    `)
	return err
}

func (c *CoreDB) SaveToken(provider, access, refresh string, expiresAt time.Time) error {
	_, err := c.db.Exec(`
        INSERT INTO tokens (provider, access_token, refresh_token, expires_at)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(provider) DO UPDATE SET
			access_token=excluded.access_token,
            refresh_token=excluded.refresh_token,
            expires_at=excluded.expires_at;
    `, provider, access, refresh, expiresAt.Unix())

	return err
}

func (c *CoreDB) GetToken(provider string) (*struct {
	AccessToken  string
	RefreshToken string
	Expires      time.Time
}, error) {
	row := c.db.QueryRow(`
        SELECT access_token, refresh_token, expires_at
        FROM tokens WHERE provider = ?;
    `, provider)

	var t struct {
		AccessToken  string
		RefreshToken string
		Expires      time.Time
	}
	var exp int64
	err := row.Scan(&t.AccessToken, &t.RefreshToken, &exp)
	if err != nil {
		return nil, err
	}

	t.Expires = time.Unix(exp, 0)
	return &t, nil
}

func (c *CoreDB) DeleteToken(provider string) error {
	_, err := c.db.Exec(`DELETE FROM tokens WHERE provider=?;`, provider)
	return err
}

func (c *CoreDB) KVGet(key string) (any, error) {
	row := c.db.QueryRow(`SELECT value FROM kv WHERE key=?;`, key)

	var value string
	err := row.Scan(&value)
	if err != nil {
		return nil, err
	}
	var nvalue any
	err = json.Unmarshal([]byte(value), &nvalue)
	if err != nil {
		return nil, err
	}

	return nvalue, nil
}

func (c *CoreDB) KVSet(key string, value any) error {

	nvalue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("KVSet(%s, %+v): failed to marshal value: %w", key, value, err)
	}
	_, nerr := c.db.Exec(`
        INSERT INTO kv (key, value)
        VALUES (?, ?)
        ON CONFLICT(key) DO UPDATE SET value=excluded.value;
    `, key, string(nvalue))

	return nerr
}

func (c *CoreDB) KVDelete(key string) error {
	_, err := c.db.Exec(`DELETE FROM kv WHERE key=?;`, key)
	return err
}

func (c *CoreDB) Close() error {
	return c.db.Close()
}
