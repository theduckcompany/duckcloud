package sqlstorage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
