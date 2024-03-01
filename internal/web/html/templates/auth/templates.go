package auth

import "html/template"

type LoginPageTmpl struct {
	UsernameContent string
	UsernameError   string

	PasswordError string
}

func (t *LoginPageTmpl) Template() string { return "auth/page_login" }

type ErrorPageTmpl struct {
	ErrorMsg  string
	RequestID string
}

func (t *ErrorPageTmpl) Template() string { return "auth/page_error" }

type ConsentPageTmpl struct {
	Username   string
	Redirect   template.URL
	ClientName string
	Scopes     []string
}

func (t *ConsentPageTmpl) Template() string { return "auth/page_consent" }

type AskMasterPasswordPageTmpl struct {
	ErrorMsg string
}

func (t *AskMasterPasswordPageTmpl) Template() string { return "auth/page_masterpassword_ask" }

type RegisterMasterPasswordPageTmpl struct {
	PasswordError string
	ConfirmError  string
}

func (t *RegisterMasterPasswordPageTmpl) Template() string {
	return "auth/page_masterpassword_register"
}
