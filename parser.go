//go:generate ragel -Z -G2 lex.rl
//go:generate goyacc parse.y
//go:generate sed "/yyS :=/a\\\tp := yylex.(*Parser)" -i y.go

package flatlang

import (
	"fmt"
	"strings"
)

func init() {
	yyErrorVerbose = true
}

type Parser struct {
	lx     *Lexer
	prev   int
	last   int
	errors []string

	Result *Node
}

func Parse(lx *Lexer) (*Parser, error) {
	p := newParser(lx)
	yyParse(p)
	if len(p.errors) == 0 {
		return p, nil
	}
	return p, fmt.Errorf("%s: %s", p.lx.file.Name(), strings.Join(p.errors, "; "))
}

func newParser(lx *Lexer) *Parser {
	return &Parser{lx: lx, prev: -1, last: len(lx.Tokens) - 1}
}

func (p *Parser) Lex(val *yySymType) int {
	if p.prev == p.last {
		return 0
	}
	p.prev++
	val.token = p.prev
	return p.lx.Tokens[p.prev].Sym
}

func (p *Parser) Error(s string) {
	p.errors = append(p.errors, p.lx.Tokens[p.prev].String()+": "+s)
}

func (p Parser) Format() string {
	return p.Result.Format(p.lx)
}
