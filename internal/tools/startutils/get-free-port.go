package startutils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort(t *testing.T) int {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)

	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)

	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
