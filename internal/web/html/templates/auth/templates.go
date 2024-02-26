package auth

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
	Redirect   string
	ClientName string
	Scopes     []string
}

func (t *ConsentPageTmpl) Template() string { return "auth/page_consent" }
