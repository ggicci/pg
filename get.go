package pg

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
)

// Get simplifies running a SELECT query which aims to find only one row of record.
//
// Usage: query a user by email, query a document by id, etc.
//
// Example:
//
//	var user = new(User)
//	var err error
//	query := pg.SQL.Select("*").From("users").Where(sq.Eq{"email": "john@example"})
//	user, err = pg.Get(ctx, user, query)
func Get[T any](ctx context.Context, v *T, query sq.SelectBuilder) (*T, error) {
	sqlstr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	err = pgxscan.Get(ctx, DB(), v, sqlstr, args...)
	return PassNotFoundError(v, err)
}
