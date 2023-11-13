package webdav

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestWebdavLitmus(t *testing.T) {
	ctx := context.Background()

	serv := startutils.NewServer(t)

	session, secret, err := serv.DavSessionsSvc.Create(ctx, &davsessions.CreateCmd{
		Name:     "litmus",
		Username: serv.User.Username(),
		UserID:   serv.User.ID(),
		Folders:  []uuid.UUID{serv.User.DefaultFolder()},
	})
	require.NoError(t, err)

	h := &Handler{
		FileSystem: serv.DFSSvc,
		Sessions:   serv.DavSessionsSvc,
		Folders:    serv.FoldersSvc,
		Files:      serv.Files,
		Logger: func(r *http.Request, err error) {
			litmus := r.Header.Get("X-Litmus")
			if len(litmus) > 19 {
				litmus = litmus[:16] + "..."
			}

			switch r.Method {
			case "COPY", "MOVE":
				dst := ""
				if u, err := url.Parse(r.Header.Get("Destination")); err == nil {
					dst = u.Path
				}
				o := r.Header.Get("Overwrite")
				t.Logf("%-20s%-10s%-30s%-30so=%-2s%v", litmus, r.Method, r.URL.Path, dst, o, err)
			default:
				t.Logf("%-20s%-10s%-30s%v", litmus, r.Method, r.URL.Path, err)
			}
		},
	}

	// The next line would normally be:
	//	http.Handle("/", h)
	// but we wrap that HTTP handler h to cater for a special case.
	//
	// The propfind_invalid2 litmus test case expects an empty namespace prefix
	// declaration to be an error. The FAQ in the webdav litmus test says:
	//
	// "What does the "propfind_invalid2" test check for?...
	//
	// If a request was sent with an XML body which included an empty namespace
	// prefix declaration (xmlns:ns1=""), then the server must reject that with
	// a "400 Bad Request" response, as it is invalid according to the XML
	// Namespace specification."
	//
	// On the other hand, the Go standard library's encoding/xml package
	// accepts an empty xmlns namespace, as per the discussion at
	// https://github.com/golang/go/issues/8068
	//
	// Empty namespaces seem disallowed in the second (2006) edition of the XML
	// standard, but allowed in a later edition. The grammar differs between
	// http://www.w3.org/TR/2006/REC-xml-names-20060816/#ns-decl and
	// http://www.w3.org/TR/REC-xml-names/#dt-prefix
	//
	// Thus, we assume that the propfind_invalid2 test is obsolete, and
	// hard-code the 400 Bad Request response that the test expects.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Litmus") == "props: 3 (propfind_invalid2)" {
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, r)

		jobErr := serv.RunnerSvc.Run(r.Context())
		if jobErr != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	t.Cleanup(ts.Close)

	t.Cleanup(func() {
		wd, err := os.Getwd()
		require.NoError(t, err, "failed to remove the logs")

		_ = os.Remove(path.Join(wd, "child.log"))
		_ = os.Remove(path.Join(wd, "debug.log"))
	})

	// Do not run the "props" and "locks" tests for now
	t.Setenv("TESTS", "basic copymove http")

	cmd := exec.Command("litmus", ts.URL, session.Username(), secret)
	require.NoError(t, err)

	stdoutBuf := bytes.NewBuffer(nil)
	cmd.Stdout = stdoutBuf

	err = cmd.Run()

	t.Log("######## RESULT ########")
	t.Log(stdoutBuf.String())
	require.NoError(t, err)
}
