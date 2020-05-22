package main

import (
	"fmt"
	"github.com/lithdew/flatlang"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	src, err := ioutil.ReadAll(os.Stdin)
	check(err)

	lx, err := flatlang.Lex(src, "")
	check(err)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	for _, tok := range lx.Tokens {
		fmt.Fprintf(w, "(input):%d:%d\t%s\t%s\n", tok.Pos, tok.End, flatlang.Repr(tok.Sym), src[tok.Pos:tok.End])
	}

	check(w.Flush())
}
