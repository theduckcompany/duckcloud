package oauthclients

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Client client model
type Client struct {
	ID             string
	Secret         string
	RedirectURI    string
	UserID         *string
	CreatedAt      time.Time
	Scopes         Scopes
	Public         bool
	SkipValidation bool
}

// GetID client id
func (c *Client) GetID() string {
	return string(c.ID)
}

// GetSecret client domain
func (c *Client) GetSecret() string {
	return c.Secret
}

// GetDomain client domain
func (c *Client) GetDomain() string {
	return c.RedirectURI
}

func (c *Client) IsPublic() bool {
	return c.Public
}

// GetUserID user id
func (c *Client) GetUserID() string {
	if c.UserID == nil {
		return ""
	}

	return *c.UserID
}

type Scopes []string

func (t Scopes) String() string {
	return strings.Join(t, ",")
}

func (t Scopes) Value() (driver.Value, error) {
	return strings.Join(t, ","), nil
}

func (t *Scopes) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("not a string")
	}

	*t = strings.Split(s, ",")

	return nil
}
