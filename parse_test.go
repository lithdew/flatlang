package flatlang

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	lx, err := LexFile("testdata/test.fbs")
	require.NoError(t, err)

	px, err := Parse(lx)
	require.NoError(t, err)

	fmt.Println(px.Format())
}
