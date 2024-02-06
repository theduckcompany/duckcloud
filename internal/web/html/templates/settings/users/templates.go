package users

import "github.com/theduckcompany/duckcloud/internal/service/users"

type ContentTemplate struct {
	IsAdmin bool
	Current *users.User
	Users   []users.User
	Error   error
}

func (t *ContentTemplate) Template() string { return "settings/users/page.tmpl" }

type RegistrationFormTemplate struct {
	Error error
}

func (t *RegistrationFormTemplate) Template() string { return "settings/users/registration-form.tmpl" }
