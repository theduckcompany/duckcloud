package storage

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		query := sq.Select("col-a", "col-b")

		query, err := PaginateSelection(query, &PaginateCmd{
			OrderBy:    []string{"col-a"},
			StartAfter: []string{"some-value"},
			Limit:      10,
		})
		require.NoError(t, err)

		raw, args, err := query.ToSql()
		require.NoError(t, err)

		assert.Equal(t, `SELECT col-a, col-b WHERE col-a = ? ORDER BY col-a LIMIT 10`, raw)
		assert.EqualValues(t, []interface{}{"some-value"}, args)
	})

	t.Run("success multi-select", func(t *testing.T) {
		query := sq.Select("col-a", "col-b")

		query, err := PaginateSelection(query, &PaginateCmd{
			OrderBy:    []string{"col-a", "col-b"},
			StartAfter: []string{"some-val-a", "some-val-b"},
			Limit:      10,
		})
		require.NoError(t, err)

		raw, args, err := query.ToSql()
		require.NoError(t, err)

		assert.Equal(t, `SELECT col-a, col-b WHERE col-a = ? AND col-b = ? ORDER BY col-a, col-b LIMIT 10`, raw)
		assert.EqualValues(t, []interface{}{"some-val-a", "some-val-b"}, args)
	})

	t.Run("without limit", func(t *testing.T) {
		query := sq.Select("col-a", "col-b")

		query, err := PaginateSelection(query, &PaginateCmd{
			OrderBy:    []string{"col-a"},
			StartAfter: []string{"some-value"},
			Limit:      -1,
		})
		require.NoError(t, err)

		raw, args, err := query.ToSql()
		require.NoError(t, err)

		assert.Equal(t, `SELECT col-a, col-b WHERE col-a = ? ORDER BY col-a`, raw)
		assert.EqualValues(t, []interface{}{"some-value"}, args)
	})
}
