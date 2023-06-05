package pg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/lann/builder"
)

// List simplifies running a SELECT query which aims to get a list of resources (rows).
//
// NOTE: `vs` is a slice value, not a pointer to a slice.
//
// Example:
//
//	var users []*User
//	err := pg.List(ctx, users, pg.SQL.Select("*").From("users"))
func List[T any](ctx context.Context, vs T, query sq.SelectBuilder, opts ...ListOption) (T, *OffsetPagination, error) {
	filteringOpts, pagingOpts, sortingOpts := CategorizedListOptions(opts...)

	if len(pagingOpts) == 0 {
		pagingOpts = []ListOption{WithOffsetPagination(NewOffsetPagination(20))}
	}
	if len(pagingOpts) > 1 {
		return vs, nil, errors.New("only one pagination option is allowed")
	}
	pagination := pagingOpts[0].(*withOffsetPaginationOption).page

	for _, opt := range filteringOpts {
		query = opt.Apply(query)
	}

	sqlstr, args, err := toCountQuery(query).ToSql()
	if err != nil {
		return vs, nil, fmt.Errorf("assemble count query: %w", err)
	}

	var total int64
	if err := DB().QueryRow(ctx, sqlstr, args...).Scan(&total); err != nil {
		return vs, nil, fmt.Errorf("count records: %w", err)
	}

	pagination.SetCountRecords(total)
	if pagination.CountRecords == 0 || pagination.Page > pagination.CountPages {
		return vs, pagination, nil // skip running query
	}

	for _, opt := range sortingOpts {
		query = opt.Apply(query)
	}
	for _, opt := range pagingOpts {
		query = opt.Apply(query)
	}

	sqlstr, args, err = query.ToSql()
	if err != nil {
		return vs, nil, fmt.Errorf("assemble query: %w", err)
	}

	err = pgxscan.Select(ctx, DB(), &vs, sqlstr, args...)
	return vs, pagination, err
}

func toCountQuery(query sq.SelectBuilder) sq.SelectBuilder {
	countQuery := builder.Delete(query, "Columns").(sq.SelectBuilder)
	countQuery = countQuery.Columns("COUNT(*)")
	return countQuery
}
