package flatlang

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLex(t *testing.T) {
	lx, err := LexFile("testdata/test.fbs")
	require.NoError(t, err)
	require.NotNil(t, lx)
}
