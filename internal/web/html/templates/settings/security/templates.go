package security

import (
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type ContentTemplate struct {
	IsAdmin        bool
	CurrentSession *websessions.Session
	WebSessions    []websessions.Session
	Devices        []davsessions.DavSession
	Spaces         map[uuid.UUID]spaces.Space
}

func (t *ContentTemplate) Template() string { return "settings/security/content.tmpl" }

type PasswordFormTemplate struct {
	Error string
}

func (t *PasswordFormTemplate) Template() string { return "settings/security/password-form.tmpl" }

type WebdavFormTemplate struct {
	Error  error
	Spaces []spaces.Space
}

func (t *WebdavFormTemplate) Template() string { return "settings/security/webdav-form.tmpl" }

type WebdavResultTemplate struct {
	Secret     string
	NewSession *davsessions.DavSession
}

func (t *WebdavResultTemplate) Template() string { return "settings/security/webdav-result.tmpl" }
