package db

import (
	"context"
	"database/sql"
	"time"
)

type QueryExecutorFn = func(ctx context.Context, db DB, query string, args ...any) (Rows, error)

// DB is an interface to a sql database. It is a wrapper for the golang sql/db builtin
type DB interface {
	Stmt
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	PingContext(ctx context.Context) error
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

// Stmt is a sql prepared statement
type Stmt interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Tx is a database transaction
type Tx interface {
	Stmt
	Commit() error
	Rollback() error
	Stmt(stmt *sql.Stmt) *sql.Stmt
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
}

// Rows is an iterator for sql.Query results
type Rows interface {
	Close() error
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Err() error
	Next() bool
	NextResultSet() bool
	Scan(dest ...any) error
}

// Row is a result for for sql.QueryRow results
type Row interface {
	Scan(dest ...any) error
}
