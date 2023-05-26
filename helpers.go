package pg

import (
	"github.com/georgysavva/scany/v2/pgxscan"
)

// PassNotFoundError swallows the `pgxscan.NotFound` error and returns nil.
func PassNotFoundError[T any](v *T, err error) (*T, error) {
	if err == nil {
		return v, err
	}

	if pgxscan.NotFound(err) {
		return nil, nil
	}

	return v, err
}
