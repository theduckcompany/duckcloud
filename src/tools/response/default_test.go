package response

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/stretchr/testify/assert"
	"github.com/unrolled/render"
)

func TestWriteError(t *testing.T) {
	resWriter := New(logger.NewNoop(), render.New())

	w := httptest.NewRecorder()

	err := errs.BadRequest(errors.New("some detailed error"), "invalid stuff")
	resWriter.WriteJSONError(w, err)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)

	assert.EqualError(t, err, "bad request: some detailed error")
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{ "message": "invalid stuff" }`, string(body))
}

func TestWriteUnhandledError(t *testing.T) {
	resWriter := New(logger.NewNoop(), render.New())

	w := httptest.NewRecorder()

	err := errors.New("some unknown error")
	resWriter.WriteJSONError(w, err)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.JSONEq(t, `{ "message": "internal error" }`, string(body))
}
