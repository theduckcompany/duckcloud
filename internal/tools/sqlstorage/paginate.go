package sqlstorage

import (
	"errors"

	sq "github.com/Masterminds/squirrel"
)

var ErrNonMatchingOrderAndStart = errors.New("OrderBy and StartAfter doesn't have the same number of arguments")

type PaginateCmd struct {
	StartAfter map[string]string
	Limit      int
}

func PaginateSelection(query sq.SelectBuilder, cmd *PaginateCmd) sq.SelectBuilder {
	if cmd == nil {
		return query
	}

	orderBy := []string{}
	for key := range cmd.StartAfter {
		orderBy = append(orderBy, key)
	}

	query = query.OrderBy(orderBy...)

	// TODO: Check that all the values in `cmd.OrderBy` are valid fields

	if len(cmd.StartAfter) > 0 {
		eqs := make(sq.Gt, len(cmd.StartAfter))

		for key, val := range cmd.StartAfter {
			eqs[key] = val
		}

		query = query.Where(eqs)
	}

	if cmd.Limit > 0 {
		query = query.Limit(uint64(cmd.Limit))
	}

	return query
}
