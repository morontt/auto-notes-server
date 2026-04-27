package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"xelbot.com/auto-notes/server/internal/constants"
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
	return dbw.QueryContext(context.Background(), query, args...)
}

func (dbw *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := dbw.db.Query(query, args...)
	dbw.logQuery(ctx, start, query, args...)

	return rows, err
}

func (dbw *DB) QueryRow(query string, args ...any) *sql.Row {
	return dbw.QueryRowContext(context.Background(), query, args...)
}

func (dbw *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := dbw.db.QueryRow(query, args...)
	dbw.logQuery(ctx, start, query, args...)

	return row
}

func (dbw *DB) Exec(query string, args ...any) (sql.Result, error) {
	return dbw.ExecContext(context.Background(), query, args...)
}

func (dbw *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := dbw.db.Exec(query, args...)
	dbw.logQuery(ctx, start, query, args...)

	return result, err
}

func (dbw *DB) Close() error {
	return dbw.db.Close()
}

func (dbw *DB) logQuery(ctx context.Context, t time.Time, query string, args ...any) {
	logContext := []any{
		"query",
		cleanQueryString(query),
		"params",
		fmt.Sprintf("%+v", args),
		"duration",
		time.Since(t),
	}

	var additional []any
	if reqID, ok := ctx.Value(constants.CtxKeyRequestID).(string); ok {
		additional = append(additional, "request_id", reqID)
	}

	if len(additional) > 0 {
		logContext = append(logContext, additional...)
	}

	dbw.logger.Debug("[SQL]", logContext...)
}

func cleanQueryString(query string) string {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = regSpaces.ReplaceAllString(query, " ")

	return strings.TrimSpace(query)
}
