package jwt

import (
	"net/http"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	t.Run("decode Authorization header success", func(t *testing.T) {
		oneHour := time.Now().Add(time.Hour)

		// Create a token
		tok := jwt.NewWithClaims(jwt.SigningMethodHS512, &generates.JWTAccessClaims{
			StandardClaims: jwt.StandardClaims{
				Audience:  "some-client-id",
				Subject:   "some-user-id",
				ExpiresAt: oneHour.Unix(),
			},
		})
		rawToken, err := tok.SignedString([]byte("some-super-secret"))
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "http://some-url", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+rawToken)

		jwtSvc := NewDefault(Config{Key: "some-super-secret"})
		token, err := jwtSvc.FetchAccessToken(req)
		require.NoError(t, err)
		assert.EqualValues(t, &AccessToken{
			ClientID: "some-client-id",
			UserID:   uuid.UUID("some-user-id"),
			Raw:      rawToken,
		}, token)
	})

	t.Run("decode path parameter success", func(t *testing.T) {
		oneHour := time.Now().Add(time.Hour)

		// Create a token
		tok := jwt.NewWithClaims(jwt.SigningMethodHS512, &generates.JWTAccessClaims{
			StandardClaims: jwt.StandardClaims{
				Audience:  "some-client-id",
				Subject:   "some-user-id",
				ExpiresAt: oneHour.Unix(),
			},
		})
		rawToken, err := tok.SignedString([]byte("some-super-secret"))
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "http://some-url?access_token="+rawToken, nil)
		require.NoError(t, err)

		jwtSvc := NewDefault(Config{Key: "some-super-secret"})
		token, err := jwtSvc.FetchAccessToken(req)
		require.NoError(t, err)
		assert.EqualValues(t, &AccessToken{
			ClientID: "some-client-id",
			UserID:   uuid.UUID("some-user-id"),
			Raw:      rawToken,
		}, token)
	})

	t.Run("decode with an invalid signature", func(t *testing.T) {
		oneHour := time.Now().Add(time.Hour)

		// Create a token
		tok := jwt.NewWithClaims(jwt.SigningMethodHS512, &generates.JWTAccessClaims{
			StandardClaims: jwt.StandardClaims{
				Audience:  "some-client-id",
				Subject:   "some-user-id",
				ExpiresAt: oneHour.Unix(),
			},
		})
		rawToken, err := tok.SignedString([]byte("Invalid Key $$$")) // Must be "some-super-token"
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "http://some-url?access_token="+rawToken, nil)
		require.NoError(t, err)

		jwtSvc := NewDefault(Config{Key: "some-super-secret"})
		token, err := jwtSvc.FetchAccessToken(req)
		assert.Nil(t, token)
		assert.EqualError(t, err, "unauthorized: invalid format\nsignature is invalid")
	})

	t.Run("decode with an expired token", func(t *testing.T) {
		oneHourAgo := time.Now().Add(-time.Hour)

		// Create a token
		tok := jwt.NewWithClaims(jwt.SigningMethodHS512, &generates.JWTAccessClaims{
			StandardClaims: jwt.StandardClaims{
				Audience:  "some-client-id",
				Subject:   "some-user-id",
				ExpiresAt: oneHourAgo.Unix(),
			},
		})
		rawToken, err := tok.SignedString([]byte("some-super-secret"))
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "http://some-url?access_token="+rawToken, nil)
		require.NoError(t, err)

		jwtSvc := NewDefault(Config{Key: "some-super-secret"})
		token, err := jwtSvc.FetchAccessToken(req)
		assert.Nil(t, token)
		assert.EqualError(t, err, "unauthorized: invalid format\ninvalid access token")
	})
}
