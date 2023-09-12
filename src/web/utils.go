package web

import (
	"net/http"

	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/web/html"
)

type AccessType int

const (
	AdminOnly AccessType = iota
	AnyUser
)

type Authenticator struct {
	webSessions websessions.Service
	users       users.Service
	html        html.Writer
}

func NewAuthenticator(webSessions websessions.Service, users users.Service, html html.Writer) *Authenticator {
	return &Authenticator{webSessions, users, html}
}

func (a *Authenticator) getUserAndSession(w http.ResponseWriter, r *http.Request, access AccessType) (*users.User, *websessions.Session, bool) {
	currentSession, err := a.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return nil, nil, true
	}

	user, err := a.users.GetByID(r.Context(), currentSession.UserID())
	if err != nil {
		a.html.WriteHTMLErrorPage(w, r, err)
		return nil, nil, true
	}

	if user == nil {
		_ = a.webSessions.Logout(r, w)
		return nil, nil, true
	}

	if access == AdminOnly && !user.IsAdmin() {
		w.Write([]byte(`<div class="alert alert-danger role="alert">Action reserved to admins</div>`))
		w.WriteHeader(http.StatusUnauthorized)
		return nil, nil, true
	}

	return user, currentSession, false
}
