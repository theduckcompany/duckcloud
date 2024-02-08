package secret

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestText(t *testing.T) {
	s1 := NewText("hello")

	t.Run("MarshalJSON", func(t *testing.T) {
		res, err := s1.MarshalJSON()
		require.NoError(t, err)

		assert.Equal(t, `"*****"`, string(res))
	})

	t.Run("MarshalText", func(t *testing.T) {
		res, err := s1.MarshalText()
		require.NoError(t, err)

		assert.Equal(t, `*****`, string(res))
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, `*****`, s1.String())
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, `*****`, s1.String())
	})

	t.Run("Raw", func(t *testing.T) {
		assert.Equal(t, "hello", s1.Raw())
	})

	t.Run("Equals", func(t *testing.T) {
		s2 := NewText("hello")

		assert.True(t, s1.Equals(s2))
		assert.True(t, s2.Equals(s1))
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		var res Text

		err := res.UnmarshalJSON([]byte(`"foobar"`))
		require.NoError(t, err)

		assert.Equal(t, "foobar", res.Raw())
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		var res Text

		err := res.UnmarshalText([]byte(`foobar`))
		require.NoError(t, err)

		assert.Equal(t, "foobar", res.Raw())
	})

	t.Run("Value", func(t *testing.T) {
		v, err := s1.Value()
		require.NoError(t, err)

		assert.Equal(t, "hello", v)
	})

	t.Run("Logvalue", func(t *testing.T) {
		assert.Implements(t, (*slog.LogValuer)(nil), s1)

		res := s1.LogValue()
		assert.Equal(t, slog.StringValue(RedactText), res)
	})

	t.Run("Scan", func(t *testing.T) {
		var res Text

		err := res.Scan("foobar")
		require.NoError(t, err)

		assert.Equal(t, "foobar", res.Raw())
	})
}
