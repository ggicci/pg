package pg

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

// Exec simplifies running a INSERT/UPDATE/DELETE query. It returns the number of rows affected.
//
// Example:
//
//	query := pg.SQL.Delete("users").Where(sq.Eq{"id": 1})
//	rowsAffected, err := pg.Exec(ctx, query)
func Exec(ctx context.Context, query sq.Sqlizer) (int64, error) {
	sqlstr, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	res, err := DB().Exec(ctx, sqlstr, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}
