package jwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/oauth2.v3/generates"
)

type Default struct {
	signature string
}

func NewDefault(signature string) *Default {
	return &Default{signature}
}

func (d *Default) getSignature() string {
	return d.signature
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
		return nil, &Error{ErrInvalidAccessToken}
	}

	// Parse and verify jwt access token
	token, err := jwt.ParseWithClaims(rawToken, &generates.JWTAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAccessToken
		}
		return []byte(d.signature), nil
	})
	if err != nil {
		return nil, &Error{fmt.Errorf("%w: %w", ErrInvalidFormat, err)}
	}

	claims, ok := token.Claims.(*generates.JWTAccessClaims)
	if !ok || !token.Valid {
		return nil, &Error{ErrInvalidAccessToken}
	}

	return &AccessToken{
		ClientID: uuid.UUID(claims.Audience),
		UserID:   uuid.UUID(claims.Subject),
		Raw:      rawToken,
	}, nil
}
