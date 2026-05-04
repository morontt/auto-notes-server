package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var regSpaces = regexp.MustCompile(`\s+`)

type DB struct {
	db     *sql.DB
	logger *slog.Logger
}

func Wrap(db *sql.DB, logger *slog.Logger) *DB {
	return &DB{
		db:     db,
		logger: logger,
	}
}

func (dbw *DB) Query(query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := dbw.db.Query(query, args...)
	dbw.logQuery(start, query, args...)

	return rows, err
}

func (dbw *DB) QueryRow(query string, args ...any) *sql.Row {
	start := time.Now()
	row := dbw.db.QueryRow(query, args...)
	dbw.logQuery(start, query, args...)

	return row
}

func (dbw *DB) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := dbw.db.Exec(query, args...)
	dbw.logQuery(start, query, args...)

	return result, err
}

func (dbw *DB) Close() error {
	return dbw.db.Close()
}

func (dbw *DB) logQuery(t time.Time, query string, args ...any) {
	dbw.logger.Debug("[SQL]", "query", cleanQueryString(query), "params", fmt.Sprintf("%+v", args), "duration", time.Since(t))
}

func cleanQueryString(query string) string {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = regSpaces.ReplaceAllString(query, " ")

	return strings.TrimSpace(query)
}
