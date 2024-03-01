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

func Test_Page_MasterPassword_Ask(t *testing.T) {
	t.Run("printPage success", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewAskMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("IsMasterKeyLoaded").Return(false).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &auth.AskMasterPasswordPageTmpl{}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/master-password/ask", nil)
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
		handler := NewAskMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("IsMasterKeyLoaded").Return(true).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/master-password/ask", nil)
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
		handler := NewAskMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("LoadMasterKeyFromPassword", mock.Anything, ptr.To(secret.NewText("some-secret"))).
			Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/master-password/ask", strings.NewReader(url.Values{
			"password": []string{"some-secret"},
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

	t.Run("postForm with an invalid password", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewAskMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("LoadMasterKeyFromPassword", mock.Anything, ptr.To(secret.NewText("some-secret"))).
			Return(errs.ErrBadRequest).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &auth.AskMasterPasswordPageTmpl{
			ErrorMsg: "invalid password",
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/master-password/ask", strings.NewReader(url.Values{
			"password": []string{"some-secret"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("postForm with a LoadMasterKeyFromPassword error", func(t *testing.T) {
		masterkeyMock := masterkey.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewAskMasterPasswordPage(htmlMock, masterkeyMock)

		masterkeyMock.On("LoadMasterKeyFromPassword", mock.Anything, ptr.To(secret.NewText("some-secret"))).
			Return(errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to load the master key from password: %w", errs.ErrInternal))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/master-password/ask", strings.NewReader(url.Values{
			"password": []string{"some-secret"},
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
