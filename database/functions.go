package database

import (
	"database/sql"
	_ "embed"
	"errors"
	_ "modernc.org/sqlite"
)

func TableExists(db *sql.DB, tableName string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
