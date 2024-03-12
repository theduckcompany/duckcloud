package sqlstorage

import (
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
)

func TestSQLTime(t *testing.T) {
	someDate := time.Date(1934, 2, 9, 32, 3, 1, 2945, time.UTC)

	sqlDate := SQLTime(someDate)

	t.Run("Time success", func(t *testing.T) {
		assert.Equal(t, someDate, sqlDate.Time())
	})

	t.Run("Value", func(t *testing.T) {
		res, err := sqlDate.Value()
		require.NoError(t, err)
		assert.Equal(t, "1934-02-10T08:03:01.000002945Z", res)
	})

	t.Run("Value", func(t *testing.T) {
		invalidDate := time.Date(-999999999999999, 0, 0, 0, 0, 0, 0, time.UTC) // invalid date
		res, err := (SQLTime)(invalidDate).Value()
		require.EqualError(t, err, "Time.MarshalText: year outside of range [0,9999]")
		assert.Nil(t, res)
	})

	t.Run("Scan success", func(t *testing.T) {
		var sqlTime SQLTime

		err := sqlTime.Scan("1934-02-10T08:03:01.000002945Z")
		require.NoError(t, err)

		assert.Equal(t, someDate, sqlTime.Time())
	})

	t.Run("Scan with an invalid type", func(t *testing.T) {
		var sqlTime SQLTime

		err := sqlTime.Scan(12)
		require.EqualError(t, err, "unsuported type: int")
	})
}

func TestIntegrationSQLTime(t *testing.T) {
	db := NewTestStorage(t)

	now := time.Now().UTC()

	t.Run("Create the table", func(t *testing.T) {
		_, err := db.Exec(`CREATE TABLE test (
		key INTEGER NOT NULL,
		time TEXT NOT NULL
		)`)
		require.NoError(t, err)
	})

	t.Run("Save the date", func(t *testing.T) {
		_, err := sq.
			Insert("test").
			Columns("key", "time").
			Values(
				42,
				ptr.To(SQLTime(now)),
			).
			RunWith(db).
			Exec()
		require.NoError(t, err)
	})

	t.Run("Fetch the date", func(t *testing.T) {
		var res SQLTime

		err := sq.
			Select("time").
			From("test").
			RunWith(db).
			Where(sq.Eq{"key": 42}).
			Scan(&res)

		require.NoError(t, err)
		assert.Equal(t, now, res.Time())
	})
}
