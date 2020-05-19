package flatlang

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

const (
	EOF rune = 0
)

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenIdent
	TokenBool
	TokenInt
	TokenFloat
	TokenString
	TokenRawString
	TokenComment
	TokenEQ
	TokenGT
	TokenGTE
	TokenLT
	TokenLTE
	TokenComma
	TokenColon
	TokenSemicolon
	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenListStart
	TokenListEnd
	TokenMapStart
	TokenMapEnd
	TokenGroupStart
	TokenGroupEnd
	TokenAND
	TokenOR
)

var TokenStr = [...]string{
	TokenError:      "Error",
	TokenEOF:        "EOF",
	TokenIdent:      "Ident",
	TokenBool:       "Bool",
	TokenInt:        "Int",
	TokenFloat:      "Float",
	TokenString:     "String",
	TokenRawString:  "RawString",
	TokenComment:    "Comment",
	TokenEQ:         "=",
	TokenGT:         ">",
	TokenGTE:        ">=",
	TokenLT:         "<",
	TokenLTE:        "<=",
	TokenComma:      ",",
	TokenColon:      ":",
	TokenSemicolon:  ";",
	TokenPlus:       "+",
	TokenMinus:      "-",
	TokenMultiply:   "*",
	TokenDivide:     "/",
	TokenListStart:  "[",
	TokenListEnd:    "]",
	TokenMapStart:   "{",
	TokenMapEnd:     "}",
	TokenGroupStart: "(",
	TokenGroupEnd:   ")",
	TokenAND:        "&",
	TokenOR:         "|",
}

func (t TokenType) String() string {
	if t < 0 || int(t) > len(TokenStr) {
		return fmt.Sprintf("Unknown(%d)", t)
	}
	return TokenStr[t]
}

type Token struct {
	Type   TokenType
	Val    string
	Line   int
	Column int
}

func (t Token) String() string {
	var b []byte
	b = strconv.AppendInt(b, int64(t.Line), 10)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(t.Column), 10)
	b = append(b, ": "...)
	b = append(b, t.Val...)
	return string(b)
}

func lower(r rune) rune { return ('a' - 'A') | r }

func isWhitespaceRune(r rune) bool {
	return uint64(1<<'\t'|1<<'\n'|1<<'\r'|1<<' ')&(1<<uint(r)) != 0
}

func isIdentRune(i int, r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) && i > 0 || r == '_' && i > 0
}

func isDecimalRune(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexadecimalRune(r rune) bool {
	return isDecimalRune(r) || (lower(r) >= 'a' && lower(r) <= 'f')
}

func decimalTypeName(prefix rune) string {
	switch prefix {
	case 'x':
		return "hexadecimal literal"
	case 'o', '0':
		return "octal literal"
	case 'b':
		return "binary literal"
	default:
		return "decimal literal"
	}
}

func lexIdent(l *Lexer, r rune) {
	r = l.next()
	for i := 1; isIdentRune(i, r); i++ {
		r = l.next()
	}
	l.backup()

	if l.current() == "true" || l.current() == "false" {
		l.emit(TokenBool)
	} else {
		l.emit(TokenIdent)
	}
}

func lexNumber(l *Lexer, r rune, dot bool) {
	r = l.next()

	prefix, invalid := EOF, EOF

	base := 10
	digit := false
	separator := false

	token := TokenError

	digits := func() {
		for {
			switch {
			case base <= 10 && isDecimalRune(r):
				digit = true
				if invalid == EOF && r >= rune(base+'0') {
					invalid = r
				}
				r = l.next()
				continue
			case base > 10 && isHexadecimalRune(r):
				digit = true
				r = l.next()
				continue
			case r == '_':
				separator = true
				r = l.next()
				continue
			}
			break
		}
	}

	if !dot {
		token = TokenInt
		if r == '0' {
			r = l.next()

			switch lower(r) {
			case 'x':
				r = l.next()
				base, prefix = 16, 'x'
			case 'o':
				r = l.next()
				base, prefix = 8, 'o'
			case 'b':
				r = l.next()
				base, prefix = 2, 'b'
			default:
				base, prefix = 8, '0'
				digit = true
			}
		}
		digits()
		if r == '.' {
			dot = true
			r = l.next()
		}
	}

	if dot {
		token = TokenFloat
		if prefix == 'o' || prefix == 'b' {
			l.errorf("invalid radix point in %s", decimalTypeName(prefix))
		}
		digits()
	}

	if !digit {
		l.errorf("%s has no digits", decimalTypeName(prefix))
	}

	e := lower(r)

	switch {
	case e == 'e' || e == 'p':
		token = TokenFloat
		base = 10
		switch {
		case e == 'e' && prefix != 0 && prefix != '0':
			l.errorf("%q exponent requires decimal mantissa", r)
		case e == 'p' && prefix != 'x':
			l.errorf("%q exponent requires hexadecimal mantissa", r)
		}
		r = l.next()
		if r == '+' || r == '-' {
			r = l.next()
		}
		digits()
		if !digit {
			l.error("exponent has no digits")
		}
	case token == TokenFloat && prefix == 'x':
		l.error("hexadecimal mantissa requires a 'p' exponent")
	case token == TokenInt && invalid != EOF:
		l.errorf("invalid digit %q in %s", invalid, decimalTypeName(prefix))
	}

	if separator && indexOfInvalidSeparator(l.current()) >= 0 {
		l.error("'_' must separate successive digits")
	}

	l.backup()
	l.emit(token)
}

func indexOfInvalidSeparator(x string) int {
	i, dec, prefix := 0, '.', ' '
	if len(x) >= 2 && x[0] == '0' {
		prefix = lower(rune(x[1]))
		if prefix == 'x' || prefix == 'o' || prefix == 'b' {
			dec = '0'
			i = 2
		}
	}
	for ; i < len(x); i++ {
		p := dec
		dec = rune(x[i])
		switch {
		case dec == '_':
			if p != '0' {
				return i
			}
		case isDecimalRune(dec) || prefix == 'x' && isHexadecimalRune(dec):
			dec = '0'
		default:
			if p == '_' {
				return i - 1
			}
			dec = '.'
		}
	}
	if dec != '_' {
		return -1
	}
	return len(x) - 1
}

func digit(r rune) int {
	if isDecimalRune(r) {
		return int(r - '0')
	}
	r = lower(r)
	if r >= 'a' && r <= 'f' {
		return int(r - 'a' + 10)
	}
	return 16
}

func lexDigits(l *Lexer, r rune, base, n int) rune {
	for n > 0 && digit(r) < base {
		r = l.next()
		n--
	}
	if n > 0 {
		l.error("invalid char escape")
	}
	return r
}

func lexEscape(l *Lexer, quote rune) rune {
	r := l.next()
	switch r {
	case quote, 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\':
		r = l.next()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		r = lexDigits(l, r, 8, 3)
	case 'x':
		r = lexDigits(l, l.next(), 16, 2)
	case 'u':
		r = lexDigits(l, l.next(), 16, 4)
	case 'U':
		r = lexDigits(l, l.next(), 16, 8)
	default:
		l.error("invalid char escape")
	}
	l.backup()
	return r
}

type Lexer struct {
	in  string
	out []Token

	eof bool

	off int
	pos int

	line    int
	lineLen int

	column    int
	columnLen int
}

func NewLexer(in string) *Lexer {
	return &Lexer{
		in:  in,
		out: make([]Token, 0, 10),

		line:    1,
		lineLen: -1,

		column:    1,
		columnLen: -1,
	}
}

func lexString(l *Lexer, quote rune) {
	l.next()

	l.skip(func(r rune) bool {
		if r == EOF || r == '\n' {
			l.error("string literal not terminated")
			return false
		}
		if r == quote {
			l.next()
			return false
		}
		if r == '\\' {
			r = lexEscape(l, quote)
		}
		return true
	})

	l.emit(TokenString)
}

func lexRawString(l *Lexer, quote rune) {
	l.next()

	l.skip(func(r rune) bool {
		if r == EOF {
			l.error("raw string literal not terminated")
			return false
		}
		if r == quote {
			l.next()
			return false
		}
		return true
	})

	l.emit(TokenRawString)
}

func (l *Lexer) Next() Token {
	for {
		if len(l.out) > 0 {
			tok := l.out[0]
			l.out = l.out[1:]
			return tok
		}

		l.skip(isWhitespaceRune)
		l.ignore()

		r := l.peek()

		if r == EOF {
			if l.pos <= l.off {
				l.emit(TokenEOF)
			} else {
				l.error("unexpected EOF")
			}
			continue
		}

		switch {
		case isIdentRune(0, r):
			lexIdent(l, r)
		case isDecimalRune(r):
			lexNumber(l, r, false)
		default:
			switch r {
			case '.':
				r = l.next()
				r = l.peek()

				if isDecimalRune(r) {
					lexNumber(l, r, true)
				}
			case '"', '\'':
				lexString(l, r)
			case '`':
				lexRawString(l, r)
			case '/':
				lexComment(l)
			case '<':
				l.next()
				if l.accept('=') {
					l.emit(TokenLTE)
				} else {
					l.emit(TokenLT)
				}
			case '>':
				l.next()
				if l.accept('=') {
					l.emit(TokenGTE)
				} else {
					l.emit(TokenGT)
				}
			case ':':
				l.next()
				l.emit(TokenColon)
			case ';':
				l.next()
				l.emit(TokenSemicolon)
			case ',':
				l.next()
				l.emit(TokenComma)
			case '=':
				l.next()
				l.emit(TokenEQ)
			case '+':
				l.next()
				l.emit(TokenPlus)
			case '-':
				l.next()
				l.emit(TokenMinus)
			case '*':
				l.next()
				l.emit(TokenMultiply)
			case '[':
				l.next()
				l.emit(TokenListStart)
			case ']':
				l.next()
				l.emit(TokenListEnd)
			case '{':
				l.next()
				l.emit(TokenMapStart)
			case '}':
				l.next()
				l.emit(TokenMapEnd)
			case '(':
				l.next()
				l.emit(TokenGroupStart)
			case ')':
				l.next()
				l.emit(TokenGroupEnd)
			case '&':
				l.next()
				l.emit(TokenAND)
			case '|':
				l.next()
				l.emit(TokenOR)
			default:
				l.next()
				l.errorf("unexpected rune '%c'", r)
			}
		}
	}
}

func lexComment(l *Lexer) {
	l.next()
	if l.accept('/') {
		l.skip(func(r rune) bool { return r != EOF && r != '\n' })
		l.emit(TokenComment)
	} else {
		l.emit(TokenDivide)
	}
}

func (l *Lexer) next() rune {
	if l.eof {
		panic("BUG: next called after eof")
	}
	if l.pos >= len(l.in) {
		l.eof = true
		return EOF
	}
	if l.in[l.pos] == '\n' {
		l.line++
		l.lineLen = l.column
		l.column = 0
	}

	r, w := utf8.DecodeRuneInString(l.in[l.pos:])

	l.pos += w
	l.column++
	l.columnLen = w

	return r
}

func (l *Lexer) backup() {
	if l.eof {
		l.eof = false
		return
	}

	if l.columnLen == -1 {
		panic("BUG: backed up too far")
	}
	l.pos -= l.columnLen
	l.column--
	l.columnLen = -1

	if l.pos < len(l.in) && l.in[l.pos] == '\n' {
		l.line--
		if l.lineLen == -1 {
			panic("BUG: backed up too many lines")
		}
		l.column = l.lineLen
		l.lineLen = -1
	}
}

func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Lexer) current() string { return l.in[l.off:l.pos] }

func (l *Lexer) ignore() { l.off = l.pos }

func (l *Lexer) skip(pred func(rune) bool) {
	for {
		r := l.next()
		if !pred(r) {
			break
		}
	}
	l.backup()
}

func (l *Lexer) accept(valid rune) bool {
	if l.next() == valid {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) emit(typ TokenType) {
	l.out = append(l.out, Token{Type: typ, Val: l.current(), Line: l.line, Column: l.column})
	l.ignore()
}

func (l *Lexer) error(val string) {
	l.out = append(l.out, Token{Type: TokenError, Val: val, Line: l.line, Column: l.column})
}

func (l *Lexer) errorf(format string, values ...interface{}) {
	l.error(fmt.Sprintf(format, values...))
}
