package secret

import (
	"encoding/base64"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKey(t *testing.T) {
	k1, err := NewKey()
	require.NoError(t, err)

	t.Run("MarshalJSON", func(t *testing.T) {
		res, err := k1.MarshalJSON()
		require.NoError(t, err)

		assert.Equal(t, `"*****"`, string(res))
	})

	t.Run("MarshalText", func(t *testing.T) {
		res, err := k1.MarshalText()
		require.NoError(t, err)

		assert.Equal(t, `*****`, string(res))
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, `*****`, k1.String())
	})

	t.Run("Base64", func(t *testing.T) {
		res, err := base64.RawStdEncoding.Strict().DecodeString(k1.Base64())
		require.NoError(t, err)

		assert.Len(t, res, KeyLength)
	})

	t.Run("Equals", func(t *testing.T) {
		k2, err := KeyFromBase64(k1.Base64())
		require.NoError(t, err)

		assert.True(t, k1.Equals(k2))
		assert.True(t, k2.Equals(k1))
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		var res Key

		err := res.UnmarshalJSON([]byte(`"lVLJtxsIkQkaiNR0QXYGH7zK9sFM4/Mfw9GwQnYGIO8"`))
		require.NoError(t, err)

		expected, err := KeyFromBase64("lVLJtxsIkQkaiNR0QXYGH7zK9sFM4/Mfw9GwQnYGIO8")
		require.NoError(t, err)

		assert.True(t, res.Equals(expected))
	})

	t.Run("UnmarshalJSON with an encoding error", func(t *testing.T) {
		var res Key

		err := res.UnmarshalJSON([]byte(`"invalid"`))
		require.EqualError(t, err, "decoding error: illegal base64 data at input byte 6")
	})

	t.Run("UnmarshalJSON with an invalid type", func(t *testing.T) {
		var res Key

		err := res.UnmarshalJSON([]byte(`32`))
		require.EqualError(t, err, "json: cannot unmarshal number into Go value of type string")
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		var res Key

		err := res.UnmarshalText([]byte(`nSHd8MGyRi6FLwjT82u4Tg7w2LGaVg3mwmYnEWmrzqM`))
		require.NoError(t, err)

		assert.NotEmpty(t, res.v)
	})

	t.Run("UnmarshalText with an decoding error", func(t *testing.T) {
		var res Key

		err := res.UnmarshalText([]byte("invalid"))
		require.EqualError(t, err, "decoding error: illegal base64 data at input byte 6")
	})

	t.Run("Value", func(t *testing.T) {
		v, err := k1.Value()
		require.NoError(t, err)

		assert.IsType(t, []byte{}, v)
		assert.Len(t, v, KeyLength)
	})

	t.Run("Logvalue", func(t *testing.T) {
		assert.Implements(t, (*slog.LogValuer)(nil), k1)

		res := k1.LogValue()
		assert.Equal(t, slog.StringValue(RedactText), res)
	})

	t.Run("Scan", func(t *testing.T) {
		var res Key

		expected, err := KeyFromBase64("lVLJtxsIkQkaiNR0QXYGH7zK9sFM4/Mfw9GwQnYGIO8")
		require.NoError(t, err)

		err = res.Scan(expected.v[:])
		require.NoError(t, err)

		assert.True(t, res.Equals(expected))
	})

	t.Run("Scan with an invalid type", func(t *testing.T) {
		var res Key

		err := res.Scan("lVLJtxsIkQkaiNR0QXYGH7zK9sFM4/Mfw9GwQnYGIO8")
		require.EqualError(t, err, "expected a []byte")
	})

	t.Run("FromRaw success", func(t *testing.T) {
		k2, err := KeyFromRaw(k1.Raw())
		require.NoError(t, err)
		assert.True(t, k2.Equals(k1))
	})

	t.Run("FromRaw with an invalid size", func(t *testing.T) {
		k2, err := KeyFromRaw([]byte("invalid key"))

		assert.Nil(t, k2)
		require.ErrorContains(t, err, "invalid key size")
	})
}
