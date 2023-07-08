package jwt

import (
	"net/http"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/go-oauth2/oauth2/v4/generates"
)

type Config struct {
	Key string `mapstructure:"key"`
}

type AccessToken struct {
	ClientID uuid.UUID
	UserID   uuid.UUID
	Raw      string
}

//go:generate mockery --name Parser
type Parser interface {
	FetchAccessToken(r *http.Request, permissions ...string) (*AccessToken, error)
	GenerateAccess() *generates.JWTAccessGenerate
}
