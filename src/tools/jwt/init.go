package jwt

import (
	"net/http"

	"github.com/Peltoche/neurone/src/tools/uuid"
)

type Config struct {
	Key string `mapstructure:"key"`
}

type AccessToken struct {
	ClientID uuid.UUID
	UserID   uuid.UUID
	Raw      string
}

type Parser interface {
	FetchAccessToken(r *http.Request, permissions ...string) (*AccessToken, error)
	getSignature() string
}