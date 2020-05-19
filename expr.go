package flatlang

import (
	"fmt"
	"strconv"
)

type PrefixParser interface {
	Parse(parser *Parser, tok Token) (Node, error)
}

type InfixParser interface {
	Parse(parser *Parser, left Node, tok Token) (Node, error)
	Precedence() int
}

var Prefix = map[TokenType]PrefixParser{
	TokenGroupStart: GroupParser(0),
	TokenIdent:      IdentParser(0),
	TokenBool:       BoolParser(0),
	TokenInt:        IntParser(0),
	TokenFloat:      FloatParser(0),
	TokenString:     StringParser(0),
	TokenRawString:  StringParser(0),

	TokenGTE:   UnaryParser(6),
	TokenGT:    UnaryParser(6),
	TokenLT:    UnaryParser(6),
	TokenLTE:   UnaryParser(6),
	TokenPlus:  UnaryParser(6),
	TokenMinus: UnaryParser(6),
}

var Infix = map[TokenType]InfixParser{
	TokenAND:      BinaryParser(1),
	TokenOR:       BinaryParser(1),
	TokenPlus:     BinaryParser(2),
	TokenMinus:    BinaryParser(2),
	TokenMultiply: BinaryParser(3),
	TokenDivide:   BinaryParser(3),
}

func parseLiteralExpr(p *Parser, precedence int) (Node, error) {
	infixPrecedence := func() int {
		infix, ok := Infix[p.current.Type]
		if !ok {
			return 0
		}
		return infix.Precedence()
	}

	current := p.current
	p.advance()

	prefix, ok := Prefix[current.Type]
	if !ok {
		return nil, fmt.Errorf("unexpected prefix token: %q", current)
	}
	left, err := prefix.Parse(p, current)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prefix token %s: %w", current, err)
	}

	for precedence < infixPrecedence() {
		current = p.current
		p.advance()

		infix, ok := Infix[current.Type]
		if !ok {
			return nil, fmt.Errorf("unexpected infix token: %q", current)
		}
		left, err = infix.Parse(p, left, current)
		if err != nil {
			return nil, fmt.Errorf("failed to parse infix token (%q): %w", current, err)
		}
	}
	return left, nil
}

type UnaryParser int

func (p UnaryParser) Parse(parser *Parser, tok Token) (Node, error) {
	right, err := parseLiteralExpr(parser, int(p))
	if err != nil {
		return nil, fmt.Errorf("unary expr: %w", err)
	}
	switch tok.Type {
	case TokenGT, TokenGTE, TokenLT, TokenLTE:
		switch right.(type) {
		case IntNode, FloatNode:
		default:
			return nil, fmt.Errorf(`unary: cannot perform op (%q) against %v`, tok.Type, right)
		}
	case TokenPlus, TokenMinus:
		switch right.(type) {
		case IntNode, FloatNode:
		default:
			return nil, fmt.Errorf(`unary: cannot perform op (%q) against val %v`, tok.Type, right)
		}
	}
	return UnaryNode{Type: tok.Type, Right: right}, nil
}

type BinaryParser int

func (p BinaryParser) Parse(parser *Parser, left Node, tok Token) (Node, error) {
	right, err := parseLiteralExpr(parser, int(p))
	if err != nil {
		return nil, fmt.Errorf("binary expr: %w", err)
	}

	_, sa := left.(StringNode)
	_, sb := right.(StringNode)

	_, ba := left.(BoolNode)
	_, bb := right.(BoolNode)

	_, ua := left.(UnaryNode)
	_, ub := right.(UnaryNode)

	_, bba := left.(BinaryNode)
	_, bbb := right.(BinaryNode)

	badBIN := func(b Node) bool {
		switch b.(BinaryNode).Type {
		case TokenAND, TokenOR:
			return true
		default:
			return false
		}
	}

	badUN := func(b Node) bool {
		switch b.(UnaryNode).Type {
		case TokenGT, TokenGTE, TokenLT, TokenLTE:
			return true
		default:
			return false
		}
	}

	switch tok.Type {
	case TokenPlus, TokenMinus, TokenMultiply, TokenDivide:
		switch {
		case sa && !sb:
			fallthrough
		case !sa && sb:
			fallthrough
		case ba || bb:
			fallthrough
		case ua && badUN(left) || ub && badUN(right):
			fallthrough
		case bba && badBIN(left) || bbb && badBIN(right):
			return nil, fmt.Errorf("binary: cannot perform op (%q) on %v and %v", tok.Type, left, right)
		}
	}

	return BinaryNode{Left: left, Type: tok.Type, Right: right}, nil
}
func (p BinaryParser) Precedence() int { return int(p) }

type GroupParser int

func (p GroupParser) Parse(parser *Parser, _ Token) (Node, error) {
	node, err := parseLiteralExpr(parser, int(p))
	if err != nil {
		return nil, fmt.Errorf("group: %w", err)
	}
	if parser.current.Type != TokenGroupEnd {
		return nil, fmt.Errorf(`group: expected ")", got %q`, parser.current.Type)
	}
	parser.advance()
	return node, nil
}

type IdentParser int

func (p IdentParser) Parse(_ *Parser, tok Token) (Node, error) {
	return IdentNode{Val: tok.Val}, nil
}

type IntParser int

func (p IntParser) Parse(_ *Parser, tok Token) (Node, error) {
	val, err := strconv.ParseInt(tok.Val, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse int: %w", err)
	}
	return IntNode{Val: val}, nil
}

type FloatParser int

func (p FloatParser) Parse(_ *Parser, tok Token) (Node, error) {
	val, err := strconv.ParseFloat(tok.Val, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse float: %w", err)
	}
	return FloatNode{Val: val}, nil
}

type StringParser int

func (p StringParser) Parse(_ *Parser, tok Token) (Node, error) {
	val, err := unquote(tok.Val)
	if err != nil {
		return nil, fmt.Errorf("failed to parse string: %w", err)
	}
	return StringNode{Val: val}, nil
}

type BoolParser int

func (p BoolParser) Parse(_ *Parser, tok Token) (Node, error) {
	switch tok.Val {
	case "true":
		return BoolNode{Val: true}, nil
	case "false":
		return BoolNode{Val: false}, nil
	default:
		return nil, fmt.Errorf("got unexpected bool value: %q", tok.Val)
	}
}
