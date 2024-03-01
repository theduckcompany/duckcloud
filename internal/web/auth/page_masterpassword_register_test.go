package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/auth"
)

func Test_Page_MasterPassword_Register(t *testing.T) {
	t.Run("printPage success", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewRegisterMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("IsMasterKeyLoaded").Return(false).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &auth.RegisterMasterPasswordPageTmpl{}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/master-password/register", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("printPage with a master key already loaded", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewRegisterMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("IsMasterKeyLoaded").Return(true).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/master-password/register", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusSeeOther, res.StatusCode)
		assert.Equal(t, "/", res.Header.Get("Location"))
	})

	t.Run("postForm success", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewRegisterMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("GenerateMasterKey", mock.Anything, ptr.To(secret.NewText("some-secret"))).
			Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/master-password/register", strings.NewReader(url.Values{
			"password": []string{"some-secret"},
			"confirm":  []string{"some-secret"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/", res.Header.Get("Location"))
	})

	t.Run("postForm with an invalid password confirmation", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewRegisterMasterPasswordPage(htmlMock, masterkeyMock)

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &auth.RegisterMasterPasswordPageTmpl{
			PasswordError: "",
			ConfirmError:  "not identical",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/master-password/register", strings.NewReader(url.Values{
			"password": []string{"some-secret"},
			"confirm":  []string{"not-the-same-secret"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("postForm with a password too short", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewRegisterMasterPasswordPage(htmlMock, masterkeyMock)

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &auth.RegisterMasterPasswordPageTmpl{
			PasswordError: "too short",
			ConfirmError:  "",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/master-password/register", strings.NewReader(url.Values{
			"password": []string{"short"},
			"confirm":  []string{"short"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("postForm with a GenerateMasterKey error", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewRegisterMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("GenerateMasterKey", mock.Anything, ptr.To(secret.NewText("some-secret"))).
			Return(errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to generate the master key: %w", errs.ErrInternal))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/master-password/register", strings.NewReader(url.Values{
			"password": []string{"some-secret"},
			"confirm":  []string{"some-secret"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
