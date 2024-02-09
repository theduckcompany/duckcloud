package users

import "github.com/theduckcompany/duckcloud/internal/service/users"

type ContentTemplate struct {
	Error   error
	Current *users.User
	Users   []users.User
	IsAdmin bool
}

func (t *ContentTemplate) Template() string { return "settings/users/page" }

type RegistrationFormTemplate struct {
	Error error
}

func (t *RegistrationFormTemplate) Template() string { return "settings/users/registration-form" }
