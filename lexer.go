package flatlang

import (
	"fmt"
	"go/token"
	"io/ioutil"
)

var fileset = token.NewFileSet()

type Token struct{ Sym, Pos, End, Prev int }

func (t Token) String() string {
	return fmt.Sprintf("%d-%d:%s", t.Pos, t.End, Repr(t.Sym))
}

type Lexer struct {
	file     *token.File
	Data     []byte
	Tokens   []Token
	Comments []Token
}

func Lex(data []byte, path string) (*Lexer, error) {
	result := newLexer(path, len(data))
	if err := lexData(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func LexFile(path string) (r *Lexer, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return Lex(data, path)
}

func newLexer(path string, size int) *Lexer {
	if path == "" {
		path = "(input)"
	}
	return &Lexer{file: fileset.AddFile(path, -1, size)}
}

func (r *Lexer) At(offset int) string {
	p := r.file.Position(r.file.Pos(offset))
	return fmt.Sprintf("%s:%d:%d: ", p.Filename, p.Line, p.Column)
}

func (r *Lexer) Last() string {
	tok := r.Tokens[len(r.Tokens)-1]
	return r.At(tok.Pos) + Repr(tok.Sym)
}

func (r *Lexer) Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("%s "+format, append([]interface{}{r.Last()}, a...))
}
