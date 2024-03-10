package sqlstorage

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		query := PaginateSelection(sq.Select("col-a", "col-b"), &PaginateCmd{
			StartAfter: map[string]string{"col-a": "some-value"},
			Limit:      10,
		})

		raw, args, err := query.ToSql()
		require.NoError(t, err)

		assert.Equal(t, `SELECT col-a, col-b WHERE col-a > ? ORDER BY col-a LIMIT 10`, raw)
		assert.EqualValues(t, []interface{}{"some-value"}, args)
	})
	t.Run("success with nil", func(t *testing.T) {
		query := PaginateSelection(sq.Select("col-a", "col-b"), nil)

		raw, args, err := query.ToSql()
		require.NoError(t, err)

		assert.Equal(t, `SELECT col-a, col-b`, raw)
		assert.EqualValues(t, []interface{}(nil), args)
	})

	t.Run("success multi-select", func(t *testing.T) {
		query := PaginateSelection(sq.Select("col-a", "col-b"), &PaginateCmd{
			StartAfter: map[string]string{
				"col-a": "some-val-a",
				"col-b": "some-val-b",
			},
			Limit: 10,
		})

		raw, args, err := query.ToSql()
		require.NoError(t, err)

		assert.Contains(t, raw, `SELECT col-a, col-b WHERE col-a > ? AND col-b > ?`)
		assert.Contains(t, raw, `ORDER BY col-`)
		assert.Contains(t, raw, `LIMIT 10`)

		assert.EqualValues(t, []interface{}{"some-val-a", "some-val-b"}, args)
	})

	t.Run("without limit", func(t *testing.T) {
		query := PaginateSelection(sq.Select("col-a", "col-b"), &PaginateCmd{
			StartAfter: map[string]string{},
			Limit:      -1,
		})

		raw, args, err := query.ToSql()
		require.NoError(t, err)

		assert.Equal(t, `SELECT col-a, col-b`, raw)
		assert.Empty(t, args)
	})
}
