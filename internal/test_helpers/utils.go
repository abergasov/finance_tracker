package testhelpers

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetFreePort generate a free tcp port for testing
func GetFreePort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)

	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, l.Close())
	}()
	return l.Addr().(*net.TCPAddr).Port
}

func SkipCIRun(t *testing.T) {
	t.Helper()
	if os.Getenv("CI_RUN") != "" {
		t.Skip("Skipping CI integration tests.")
	}
}
