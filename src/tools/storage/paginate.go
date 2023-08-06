package storage

import (
	"errors"

	sq "github.com/Masterminds/squirrel"
)

var (
	ErrNonMatchingOrderAndStart = errors.New("OrderBy and StartAfter doesn't have the same number of arguments")
)

type PaginateCmd struct {
	OrderBy    []string
	StartAfter []string
	Limit      int
}

func PaginateSelection(query sq.SelectBuilder, cmd *PaginateCmd) (sq.SelectBuilder, error) {
	query = query.OrderBy(cmd.OrderBy...)

	// TODO: Check that all the values in `cmd.OrderBy` are valid fields

	if len(cmd.OrderBy) != len(cmd.StartAfter) {
		return sq.SelectBuilder{}, ErrNonMatchingOrderAndStart
	}

	if len(cmd.StartAfter) > 0 {
		eqs := make(sq.Eq, len(cmd.OrderBy))

		for idx, elem := range cmd.OrderBy {
			eqs[elem] = cmd.StartAfter[idx]
		}

		query = query.Where(eqs)
	}

	if cmd.Limit > 0 {
		query = query.Limit(uint64(cmd.Limit))
	}

	return query, nil
}
