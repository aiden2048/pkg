package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aiden2048/pkg/frame"
)

type Ck struct {
	platformId int32
	db         string
	tb         string
}

func NewCk(db, tb string) *Ck {
	return &Ck{
		platformId: frame.GetPlatformId(),
		db:         db,
		tb:         tb,
	}
}

// SetPlatformId sets the platform ID for the connection lookup
func (c *Ck) SetPlatformId(platId int32) {
	c.platformId = platId
}

// GetConn returns the *sql.DB instance
func (c *Ck) GetConn() (*sql.DB, error) {
	db := GetCkDb(c.platformId)
	if db == nil {
		return nil, fmt.Errorf("clickhouse db not found for platform %d", c.platformId)
	}
	return db, nil
}

// Exec executes a query without returning any rows.
func (c *Ck) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db, err := c.GetConn()
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT.
func (c *Ck) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := c.GetConn()
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (c *Ck) QueryRow(ctx context.Context, query string, args ...interface{}) (*sql.Row, error) {
	db, err := c.GetConn()
	if err != nil {
		return nil, err
	}
	return db.QueryRowContext(ctx, query, args...), nil
}
