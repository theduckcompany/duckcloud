package jwt

import (
	"net/http"

	"github.com/Peltoche/neurone/src/tools/uuid"
)

type AccessToken struct {
	ClientID uuid.UUID
	UserID   uuid.UUID
	Raw      string
}

type Parser interface {
	FetchAccessToken(r *http.Request, permissions ...string) (*AccessToken, error)
	getSignature() string
}
