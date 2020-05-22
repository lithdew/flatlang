package flatlang

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Stack [][2]int

func (s *Stack) Push(a, b int) {
	*s = append(*s, [2]int{a, b})
}

func (s *Stack) Pop() (int, int) {
	val := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return val[0], val[1]
}

func Repr(sym int) string {
	if sym >= yyPrivate-1 && sym < yyPrivate+len(yyToknames) {
		return yyToknames[sym-yyPrivate+1]
	}
	return fmt.Sprintf("%c", sym)
}

func Unquote(s string) (string, error) {
	n := len(s)
	if n < 2 {
		return "", strconv.ErrSyntax
	}
	quote := s[0]
	if quote != s[n-1] {
		return "", strconv.ErrSyntax
	}
	s = s[1 : n-1]

	if quote == '`' {
		if strings.ContainsRune(s, '`') {
			return "", strconv.ErrSyntax
		}
		if strings.ContainsRune(s, '\r') {
			buf := make([]byte, 0, len(s)-1)
			for i := 0; i < len(s); i++ {
				if s[i] != '\r' {
					buf = append(buf, s[i])
				}
			}
			return string(buf), nil
		}
		return s, nil
	}

	if quote != '"' && quote != '\'' {
		return "", strconv.ErrSyntax
	}

	if strings.ContainsRune(s, '\n') {
		return "", strconv.ErrSyntax
	}

	if !strings.ContainsRune(s, '\\') && !strings.ContainsRune(s, rune(quote)) {
		if utf8.ValidString(s) {
			return s, nil
		}
	}

	var tmp [utf8.UTFMax]byte

	buf := make([]byte, 0, 3*len(s)/2)
	for len(s) > 0 {
		c, multibyte, ss, err := strconv.UnquoteChar(s, quote)
		if err != nil {
			return "", err
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(tmp[:], c)
			buf = append(buf, tmp[:n]...)
		}
	}

	return string(buf), nil
}
