package flatlang

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseLiteralExpr(t *testing.T) {
	S := func(s string) *string { return &s }

	cases := []struct {
		src      string
		expected *string
	}{
		{
			src:      `>=100|"test"|<=500"test"1 2 3`,
			expected: S(`(((>=100) | "test") | (<=500))`),
		},
		{
			src:      `"test"|<30.4`,
			expected: S(`("test" | (<30.4))`),
		},
		{
			src:      `"hello"|(world + (3.4 + 4.9)) test`,
			expected: S(`("hello" | (world + (3.4 + 4.9)))`),
		},
		{
			src:      `40 + 100 * 32 / 40`,
			expected: S(`(40 + ((100 * 32) / 40))`),
		},
		{
			src:      `test(-1.3 + 44)`,
			expected: S(`test`),
		},
		{
			src:      `test - 1.3 + 44`,
			expected: S(`((test - 1.3) + 44)`),
		},
		{src: `40 + 100 * 32 / <=40`},
		{src: `"hello"|(world + (3.4 & "test")) test`},
		{src: `>=|`},
		{src: `>=&`},
		{src: `>"test"`},
		{src: `<false`},
		{src: `>=100&<>`},
	}

	for _, test := range cases {
		p := NewParser(NewLexer(test.src))

		actual, err := parseLiteralExpr(p, 0)
		if test.expected != nil {
			require.NoError(t, err)
			require.EqualValues(t, *test.expected, actual.Repr())
			continue
		}
		require.Error(t, err)
	}
}
