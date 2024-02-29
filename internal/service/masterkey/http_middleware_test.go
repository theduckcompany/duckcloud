package masterkey

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

func Test_HTTP_Middleware(t *testing.T) {
	t.Run("call any endpoint with a loaded master key", func(t *testing.T) {
		htmlMock := html.NewMockWriter(t)
		svcMock := NewMockService(t)

		mid := NewHTTPMiddleware(svcMock, htmlMock)

		nextCalled := false
		handler := mid.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusTeapot)
		}))

		svcMock.On("IsMasterKeyLoaded").Return(true).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo", nil)

		handler.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusTeapot, res.StatusCode)
		assert.True(t, nextCalled)
	})

	t.Run("call any endpoint with a master key neither loaded nor registered", func(t *testing.T) {
		htmlMock := html.NewMockWriter(t)
		svcMock := NewMockService(t)

		mid := NewHTTPMiddleware(svcMock, htmlMock)

		nextCalled := false
		handler := mid.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true }))

		svcMock.On("IsMasterKeyLoaded").Return(false).Once()
		svcMock.On("IsMasterKeyRegistered", mock.Anything).Return(false, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo", nil)

		handler.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusSeeOther, res.StatusCode)
		assert.Equal(t, "/master-password/register", res.Header.Get("Location"))
		assert.False(t, nextCalled)
	})

	t.Run("call any endpoint with a master key registered but not loaded", func(t *testing.T) {
		htmlMock := html.NewMockWriter(t)
		svcMock := NewMockService(t)

		mid := NewHTTPMiddleware(svcMock, htmlMock)

		nextCalled := false
		handler := mid.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true }))

		svcMock.On("IsMasterKeyLoaded").Return(false).Once()
		svcMock.On("IsMasterKeyRegistered", mock.Anything).Return(true, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo", nil)

		handler.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusSeeOther, res.StatusCode)
		assert.Equal(t, "/master-password/ask", res.Header.Get("Location"))
		assert.False(t, nextCalled)
	})

	t.Run("call any endpoint with IsMasterKeyRegistered error", func(t *testing.T) {
		htmlMock := html.NewMockWriter(t)
		svcMock := NewMockService(t)

		mid := NewHTTPMiddleware(svcMock, htmlMock)

		nextCalled := false
		handler := mid.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true }))

		svcMock.On("IsMasterKeyLoaded").Return(false).Once()
		svcMock.On("IsMasterKeyRegistered", mock.Anything).Return(false, errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to check if the master key is registered: %w", errs.ErrInternal)).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo", nil)

		handler.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.False(t, nextCalled)
	})

	t.Run("call ask endpoint with a master key not loaded", func(t *testing.T) {
		htmlMock := html.NewMockWriter(t)
		svcMock := NewMockService(t)

		mid := NewHTTPMiddleware(svcMock, htmlMock)

		nextCalled := false
		handler := mid.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusTeapot)
		}))

		svcMock.On("IsMasterKeyLoaded").Return(false).Once()
		svcMock.On("IsMasterKeyRegistered", mock.Anything).Return(true, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/master-password/ask", nil)

		handler.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusTeapot, res.StatusCode)
		assert.True(t, nextCalled)
	})

	t.Run("call register endpoint with a master key neither loaded nor registered", func(t *testing.T) {
		htmlMock := html.NewMockWriter(t)
		svcMock := NewMockService(t)

		mid := NewHTTPMiddleware(svcMock, htmlMock)

		nextCalled := false
		handler := mid.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusTeapot)
		}))

		svcMock.On("IsMasterKeyLoaded").Return(false).Once()
		svcMock.On("IsMasterKeyRegistered", mock.Anything).Return(false, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/master-password/register", nil)

		handler.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusTeapot, res.StatusCode)
		assert.True(t, nextCalled)
	})
}
