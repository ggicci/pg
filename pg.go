package pg

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// SQL is a statement builder with PostgreSQL dialect enabled.
	// Usage:
	//    query := SQL.Select("*").From("users")....
	//    query := SQL.Update("users").Set("name", "John")....
	SQL = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	pool *pgxpool.Pool
)

// Init initializes the database connection pool, using the given connection string.
// See `pgxpool.New` for more details about the format of the connection string.
func Init(ctx context.Context, connString string) (err error) {
	pool, err = pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("pgxpool.New failed: %w", err)
	}
	return pool.Ping(context.Background())
}

// DB returns the database connection pool.
func DB() *pgxpool.Pool {
	return pool
}
