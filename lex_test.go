package flatlang

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestLex(t *testing.T) {
	lx, err := LexFile("testdata/test.fbs")
	require.NoError(t, err)
	require.NotNil(t, lx)
}

func BenchmarkLex(b *testing.B) {
	src, err := ioutil.ReadFile("testdata/test.fbs")
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := lexData(src, newLexer("", len(src))); err != nil {
			b.Fatal(err)
		}
	}
}
