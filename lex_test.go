package flatlang

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path/filepath"
	"testing"
)

const DIR = "testdata"

func TestLex(t *testing.T) {
	files, err := ioutil.ReadDir(DIR)
	require.NoError(t, err)

	for _, file := range files {
		src, err := ioutil.ReadFile(filepath.Join(DIR, file.Name()))
		require.NoError(t, err)

		l := NewLexer(string(src))

		for token := l.Next(); token.Type != TokenEOF; token = l.Next() {
			if token.Type == TokenComment {
				continue
			}
		}
	}
}
