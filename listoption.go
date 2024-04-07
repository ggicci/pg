package pg

import (
	sq "github.com/Masterminds/squirrel"
)

// ListOption is a function that applies a condition to a query.
type ListOption interface {
	Apply(sq.SelectBuilder) sq.SelectBuilder
}

// ListOptionFunc is an adapter to allow the use of ordinary functions as ListOption.
type ListOptionFunc func(sq.SelectBuilder) sq.SelectBuilder

func (f ListOptionFunc) Apply(sb sq.SelectBuilder) sq.SelectBuilder {
	return f(sb)
}

// With returns a ListOption that applies the given condition to the query.
//
// When the given value is empty, the returned ListOption is a no-op.
//
// When the given value is a single value, it will be equivalent to a simple
// equality condition.
//
// When the value is a slice, it will be expanded to a list of OR conditions.
// Which is equivalent to an IN statement.
func With[T any](columnName string, value ...T) ListOption {
	return ListOptionFunc(func(sb sq.SelectBuilder) sq.SelectBuilder {
		// noop
		if len(value) == 0 {
			return sb
		}

		// columnName = value
		if len(value) == 1 {
			return sb.Where(sq.Eq{columnName: value})
		}

		// columnName IN (value...)
		var cond sq.Or
		for _, item := range value {
			cond = append(cond, sq.Eq{columnName: item})
		}
		return sb.Where(cond)
	})
}

// Without works like With, but it negates the condition. See With for more details.
func Without[T any](columnName string, value ...T) ListOption {
	return ListOptionFunc(func(sb sq.SelectBuilder) sq.SelectBuilder {
		// noop
		if len(value) == 0 {
			return sb
		}

		// columnName <> value
		if len(value) == 1 {
			return sb.Where(sq.NotEq{columnName: value})
		}

		// columnName NOT IN (value...)
		var cond sq.And
		for _, item := range value {
			cond = append(cond, sq.NotEq{columnName: item})
		}
		return sb.Where(cond)
	})
}

type withSortByOption struct {
	columnName string
	direction  string // "asc" or "desc"
}

func (o *withSortByOption) Apply(sb sq.SelectBuilder) sq.SelectBuilder {
	return sb.OrderBy(o.columnName + " " + o.direction)
}

// WithSortBy returns a ListOption that sorts the result by the given column name and sort direction.
// The direction must be either "asc" or "desc".
func WithSortBy(columnName, direction string) ListOption {
	return &withSortByOption{columnName, direction}
}

type withOffsetPaginationOption struct {
	page *OffsetPagination
}

func (o *withOffsetPaginationOption) Apply(sb sq.SelectBuilder) sq.SelectBuilder {
	return sb.Limit(uint64(o.page.Limit())).Offset(uint64(o.page.Offset()))
}

// WithOffsetPagination returns a ListOption that limits the result to the given page.
func WithOffsetPagination(pagination *OffsetPagination) ListOption {
	page := new(OffsetPagination)
	*page = *pagination
	return &withOffsetPaginationOption{page}
}

// IsPaginationOption returns true if the given ListOption is used for pagination.
func IsPaginationOption(opt ListOption) bool {
	_, ok := opt.(*withOffsetPaginationOption)
	return ok
}

// IsSortingOption returns true if the given ListOption is used for limiting the result.
func IsSortingOption(opt ListOption) bool {
	_, ok := opt.(*withSortByOption)
	return ok
}

// CategorizedListOptions categorizes the given ListOptions into types of filtering, paging, and sorting.
func CategorizedListOptions(opts ...ListOption) (filtering, paging, sorting []ListOption) {
	for _, opt := range opts {
		if IsPaginationOption(opt) {
			paging = append(paging, opt)
		} else if IsSortingOption(opt) {
			sorting = append(sorting, opt)
		} else {
			filtering = append(filtering, opt)
		}
	}
	return
}
