package flatlang

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEval(t *testing.T) {
	src := []byte("hi = 123 + 4.0; there = `this is a ${hi + 5} test`;")

	lx, err := Lex(src, "")
	require.NoError(t, err)

	px, err := Parse(lx)
	require.NoError(t, err)

	fmt.Printf("Evaluating %q.\n\n", src[:len(src)-1])

	res, err := Eval(lx, px.Result)
	require.NoError(t, err)

	fmt.Println()

	spew.Dump(res)
}
