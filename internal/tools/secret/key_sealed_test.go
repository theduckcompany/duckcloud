package secret

import (
	"encoding/base64"
	"log/slog"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSealedKey(t *testing.T) {
	masterKey, err := NewKey()
	require.NoError(t, err)

	k1, err := NewKey()
	require.NoError(t, err)

	var sk1 *SealedKey

	t.Run("SealKey", func(t *testing.T) {
		var err error

		sk1, err = SealKey(masterKey, k1)
		require.NoError(t, err)
		assert.NotEmpty(t, sk1)
	})

	t.Run("SealedKey != key", func(t *testing.T) {
		assert.NotEqual(t, k1, sk1)
	})

	t.Run("Open key", func(t *testing.T) {
		res, err := sk1.Open(masterKey)
		require.NoError(t, err)
		assert.True(t, k1.Equals(res))
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		res, err := sk1.MarshalJSON()
		require.NoError(t, err)

		assert.Equal(t, `"*****"`, string(res))
	})

	t.Run("MarshalText", func(t *testing.T) {
		res, err := sk1.MarshalText()
		require.NoError(t, err)

		assert.Equal(t, `*****`, string(res))
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, `*****`, sk1.String())
	})

	t.Run("Base64", func(t *testing.T) {
		res, err := base64.RawStdEncoding.Strict().DecodeString(sk1.Base64())
		require.NoError(t, err)

		assert.Len(t, res, SealedKeyLength)
	})

	t.Run("Equals", func(t *testing.T) {
		sk2, err := SealedKeyFromBase64(sk1.Base64())
		require.NoError(t, err)

		assert.True(t, sk1.Equals(sk2))
		assert.True(t, sk2.Equals(sk1))
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		var res SealedKey

		err := res.UnmarshalJSON([]byte(`"KU5eVc37f3377a99sQ/Q/xyBJ6anzn65exn1WYEgljoK1u3CVZwiGvPgNLEJQwMgHPWw+H0KiaK/yLEuEipceOKT9cswzVUm"`))
		require.NoError(t, err)

		expected, err := SealedKeyFromBase64("KU5eVc37f3377a99sQ/Q/xyBJ6anzn65exn1WYEgljoK1u3CVZwiGvPgNLEJQwMgHPWw+H0KiaK/yLEuEipceOKT9cswzVUm")
		require.NoError(t, err)

		assert.True(t, res.Equals(expected))
	})

	t.Run("UnmarshalJSON with an encoding error", func(t *testing.T) {
		var res SealedKey

		err := res.UnmarshalJSON([]byte(`"invalid"`))
		require.EqualError(t, err, "decoding error: illegal base64 data at input byte 6")
	})

	t.Run("UnmarshalJSON with an invalid type", func(t *testing.T) {
		var res SealedKey

		err := res.UnmarshalJSON([]byte(`32`))
		require.EqualError(t, err, "json: cannot unmarshal number into Go value of type string")
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		var res SealedKey

		err := res.UnmarshalText([]byte(`a7g0dYQdk7DexmG5Nsal1O9gMUPmxo5zfpyr6U4Mdvo/QQkyiJHmpRnYuI7IapGtlcvlbxbkySXYNlw2HZhRqQvLeAgjvbfd`))
		require.NoError(t, err)

		assert.NotEmpty(t, res.v)
	})

	t.Run("UnmarshalText with an decoding error", func(t *testing.T) {
		var res SealedKey

		err := res.UnmarshalText([]byte("invalid"))
		require.EqualError(t, err, "decoding error: illegal base64 data at input byte 6")
	})

	t.Run("Value", func(t *testing.T) {
		v, err := sk1.Value()
		require.NoError(t, err)

		assert.IsType(t, []byte{}, v)
		assert.Len(t, v, SealedKeyLength)
	})

	t.Run("Logvalue", func(t *testing.T) {
		assert.Implements(t, (*slog.LogValuer)(nil), sk1)

		res := sk1.LogValue()
		assert.Equal(t, slog.StringValue(RedactText), res)
	})

	t.Run("Scan", func(t *testing.T) {
		var res SealedKey

		expected, err := SealedKeyFromBase64("BKgSe9jPaIGQWn7EZd+44BduGVgvZyZ23rvAWSvEKJ02mwDzerGwlltpsVbDMWI2N+XimZLXKKX84TnIbF2XXKPU7V/tY4pz")
		require.NoError(t, err)

		err = res.Scan(expected.v[:])
		require.NoError(t, err)

		assert.True(t, res.Equals(expected))
	})

	t.Run("Scan with an invalid type", func(t *testing.T) {
		var res SealedKey

		err := res.Scan("BKgSe9jPaIGQWn7EZd+44BduGVgvZyZ23rvAWSvEKJ02mwDzerGwlltpsVbDMWI2N+XimZLXKKX84TnIbF2XXKPU7V/tY4pz")
		require.EqualError(t, err, "expected a []byte")
	})

	t.Run("SeaKeyWithEnclave and OpenKeyWithEnclave", func(t *testing.T) {
		mk, err := NewKey()
		require.NoError(t, err)

		enclave := memguard.NewEnclave(mk.Raw())

		seal, err := SealKeyWithEnclave(enclave, k1)
		require.NoError(t, err)

		res, err := seal.OpenWithEnclave(enclave)
		require.NoError(t, err)
		assert.True(t, res.Equals(k1))
	})
}
