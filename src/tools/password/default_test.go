package password

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptPassword(t *testing.T) {
	ctx := context.Background()

	t.Run("Encrypt/Compare success", func(t *testing.T) {
		password := &BcryptPassword{}

		hashed, err := password.Encrypt(ctx, "some-password")
		require.NoError(t, err)
		require.NotEqual(t, hashed, "some-password")

		err = password.Compare(ctx, hashed, "some-password")
		assert.NoError(t, err)
	})

	t.Run("Encrypt with a password too long", func(t *testing.T) {
		password := &BcryptPassword{}

		hashed, err := password.Encrypt(ctx, "some-very-very-very-very-very-very-very-very-very-very-very-very----very-very-very-very-long-password")
		assert.EqualError(t, err, "bcrypt: password length exceeds 72 bytes")
		assert.Empty(t, hashed)
	})

	t.Run("Decrypte with a no base64 string", func(t *testing.T) {
		password := &BcryptPassword{}

		err := password.Compare(ctx, "not a hex string#", "some-password")
		assert.EqualError(t, err, "failed to decode the password: illegal base64 data at input byte 3")
	})
}
