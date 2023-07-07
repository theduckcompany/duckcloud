package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/golang-jwt/jwt"
)

var (
	ErrInvalidAccessToken = fmt.Errorf("invalid access token")
	ErrMissingAccessToken = fmt.Errorf("missing access token")
	ErrInvalidFormat      = fmt.Errorf("invalid format")
)

type Default struct {
	signature string
}

func NewDefault(cfg Config) *Default {
	return &Default{cfg.Key}
}

func (d *Default) FetchAccessToken(r *http.Request, permissions ...string) (*AccessToken, error) {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "
	rawToken := ""

	if auth != "" && strings.HasPrefix(auth, prefix) {
		rawToken = auth[len(prefix):]
	} else {
		rawToken = r.FormValue("access_token")
	}

	if rawToken == "" {
		return nil, errs.Unauthorized(ErrInvalidAccessToken, "invalid access token")
	}

	// Parse and verify jwt access token
	token, err := jwt.ParseWithClaims(rawToken, &generates.JWTAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAccessToken
		}
		return []byte(d.signature), nil
	})
	if err != nil {
		return nil, errs.Unauthorized(errors.Join(ErrInvalidFormat, err), "invalid access token")
	}

	claims, ok := token.Claims.(*generates.JWTAccessClaims)
	if !ok || !token.Valid {
		return nil, errs.Unauthorized(ErrInvalidAccessToken, "invalid access token")
	}

	return &AccessToken{
		ClientID: uuid.UUID(claims.Audience),
		UserID:   uuid.UUID(claims.Subject),
		Raw:      rawToken,
	}, nil
}

func (d *Default) GenerateAccess() *generates.JWTAccessGenerate {
	return generates.NewJWTAccessGenerate("", []byte(d.signature), jwt.SigningMethodHS512)
}
