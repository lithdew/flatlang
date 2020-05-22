package flatlang

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEval(t *testing.T) {
	src := []byte(`hello = items 'prefix'; hello '1'; hello '2';`)

	lx, err := Lex(src, "")
	require.NoError(t, err)

	px, err := Parse(lx)
	require.NoError(t, err)

	fmt.Printf("Evaluating %q.\n\n", src[:len(src)-1])

	ex := NewEval(lx)

	printFn := func(params ...interface{}) {
		fmt.Println(params...)
	}
	require.NoError(t, ex.register("print", printFn))

	var allItems []string
	itemsFn := func(items ...string) {
		allItems = append(allItems, items...)
	}
	require.NoError(t, ex.register("items", itemsFn))

	res, err := ex.eval(px.Result)
	require.NoError(t, err)
	_ = res

	spew.Dump(allItems)
}
