package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// DatabaseSource is a source of pgx connections.
type DatabaseSource interface {
	Connect(context.Context) (*pgx.Conn, error)
}

// DSNSource is a pgx source from DSN string.
type DSNSource string

// Connect implements DatabaseSource.
func (dsn DSNSource) Connect(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, (string)(dsn))
}
