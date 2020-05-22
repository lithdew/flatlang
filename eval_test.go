package flatlang

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCurry(t *testing.T) {
	var report []string

	fn := func(items ...string) error {
		if len(items) == 0 {
			return errors.New("must specify at least one string param to 'item'")
		}
		report = append(report, items...)
		return nil
	}

	src := []byte(`hello = items 'test' > items 'test_two'; items 'start' > hello > items 'end' > hello '1';`)

	lx, err := Lex(src, "")
	require.NoError(t, err)

	px, err := Parse(lx)
	require.NoError(t, err)

	ex := NewEval(lx)

	require.NoError(t, ex.RegisterBuiltin("items", fn))

	_, err = ex.Eval(px.Result)
	require.NoError(t, err)

	require.EqualValues(t, []string{"start", "test", "test_two", "end", "test", "test_two", "1"}, report)
}

func TestEval(t *testing.T) {
	src := []byte(`dbg = print '[A]' > print '[B]'; print 'start' > dbg 'A test message!' > print 'end'; print (1 + 2 - 3) 34.5e4;`)

	lx, err := Lex(src, "")
	require.NoError(t, err)

	px, err := Parse(lx)
	require.NoError(t, err)

	fmt.Printf("Evaluating %q.\n\n", src[:len(src)-1])

	ex := NewEval(lx)

	printFn := func(params ...interface{}) {
		fmt.Println(params...)
	}
	require.NoError(t, ex.RegisterBuiltin("print", printFn))

	res, err := ex.Eval(px.Result)
	require.NoError(t, err)

	fmt.Println()

	spew.Dump(res)
}
