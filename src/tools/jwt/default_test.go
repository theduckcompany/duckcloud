package jwt

import (
	"net/http"
	"testing"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	jwtSvc := NewDefault(Config{Key: "A very bad key"})

	req, err := http.NewRequest(http.MethodGet, "http://some-url", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJuZXVyb25lLXdlYi11aSIsImV4cCI6MTY4OTcxMzM4OSwic3ViIjoiOGYwNmI5MGMtNWM4MS00Nzc5LWE5MjctMTU3MmNmYjIwNTg2In0.2PZIl9yD8bQmJnYg1kIQCfonEn9XL3jDCuCTzepWxocF1LRpB8MHe6fzNMKR-AVfT-Nqv-MMKIF9sKL1ki2dtA")

	token, err := jwtSvc.FetchAccessToken(req)
	require.NoError(t, err)
	assert.EqualValues(t, &AccessToken{
		ClientID: "neurone-web-ui",
		UserID:   uuid.UUID("8f06b90c-5c81-4779-a927-1572cfb20586"),
		Raw:      "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJuZXVyb25lLXdlYi11aSIsImV4cCI6MTY4OTcxMzM4OSwic3ViIjoiOGYwNmI5MGMtNWM4MS00Nzc5LWE5MjctMTU3MmNmYjIwNTg2In0.2PZIl9yD8bQmJnYg1kIQCfonEn9XL3jDCuCTzepWxocF1LRpB8MHe6fzNMKR-AVfT-Nqv-MMKIF9sKL1ki2dtA",
	}, token)
}
