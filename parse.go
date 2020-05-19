package flatlang

import (
	"errors"
	"fmt"
)

type Parser struct {
	lexer   *Lexer
	current Token
	peek    Token
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}
	p.advance()
	p.advance()
	return p
}

func (p *Parser) Parse() (Program, error) {
	var program Program
	for p.current.Type != TokenEOF {
		if p.current.Type == TokenComment {
			p.advance()
			continue
		}
		stmt, err := parseStmt(p)
		if err != nil {
			return program, fmt.Errorf("failed to parse stmt in program: %w", err)
		}
		program.Stmts = append(program.Stmts, stmt)
	}
	return program, nil
}

func (p *Parser) advance() {
	p.current = p.peek
	p.peek = p.lexer.Next()
}

func parseStmt(p *Parser) (Stmt, error) {
	var stmt Stmt

	if p.current.Type != TokenIdent {
		return stmt, fmt.Errorf("expected stmt to begin with <ident>, got %q", p.current)
	}

	stmt.Type = StmtTypeCall

	if p.peek.Type == TokenEQ {
		stmt.Type = StmtTypeAssign
		stmt.Name = p.current.Val
		p.advance() // <ident>
		p.advance() // "="
	}

	// If there's _only_ a single extraneous ">", just skip it.

	if p.current.Type == TokenGT {
		p.advance()
	}

	exprs, err := parseExprs(p)
	if err != nil {
		return stmt, fmt.Errorf("failed to parse stmt exprs: %w", err)
	}

	stmt.Exprs = exprs

	return stmt, nil
}

func parseExprs(p *Parser) ([]Expr, error) {
	var exprs []Expr
	for p.current.Type != TokenEOF && p.current.Type != TokenSemicolon {
		expr, err := parseExpr(p)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expressions: %w", err)
		}
		exprs = append(exprs, expr)
	}
	p.advance()
	return exprs, nil
}

func parseExpr(p *Parser) (Expr, error) {
	first := p.current

	var expr Expr
	for p.current.Type != TokenEOF && p.current.Type != TokenGT && p.current.Type != TokenSemicolon {
		switch p.current.Type {
		case TokenListStart: // "["
			list, err := parseList(p)
			if err != nil {
				return expr, fmt.Errorf("failed to parse list: %w", err)
			}
			expr.Nodes = append(expr.Nodes, list)
		case TokenMapStart: // "{"
			node, err := parseMap(p)
			if err != nil {
				return expr, fmt.Errorf("failed to parse map: %w", err)
			}
			expr.Nodes = append(expr.Nodes, node)
		default:
			lit, err := parseLiteralExpr(p, 0)
			if err != nil {
				return expr, fmt.Errorf("failed to parse literal expr: %w", err)
			}
			expr.Nodes = append(expr.Nodes, lit)
		}
	}
	if p.current.Type == TokenGT {
		p.advance()
	}
	if first.Type != TokenIdent && len(expr.Nodes) > 1 {
		return expr, errors.New("expr may only have parameters if it is a call to a method")
	}
	return expr, nil
}

func parseMap(p *Parser) (MapNode, error) {
	var node MapNode

	if p.current.Type != TokenMapStart {
		return node, errors.New(`map does not start with "{"`)
	}

	p.advance() // "{"

	for p.current.Type != TokenEOF && p.current.Type != TokenMapEnd {
		var field Field

		if p.current.Type != TokenIdent {
			return node, fmt.Errorf("map field must be an ident, but got %q", p.current)
		}

		field.Key = p.current.Val

		p.advance() // key: <ident>

		if p.current.Type != TokenColon {
			return node, fmt.Errorf(`expected ":" after map field ident key, but got %q`, p.peek)
		}

		p.advance() // ":"

		val, err := parseLiteralExpr(p, 0)
		if err != nil {
			return node, fmt.Errorf(`failed to parse map field literal expr value: %w`, err)
		}

		field.Val = val

		if p.current.Type != TokenComma && p.current.Type != TokenMapEnd {
			return node, fmt.Errorf(`expected comma or "}" after map field, but got %q`, p.current)
		}

		if p.peek.Type == TokenMapEnd {
			return node, errors.New(`extraneous "}" at end of map`)
		}

		node.Fields = append(node.Fields, field)

		if p.current.Type == TokenMapEnd {
			break
		}

		p.advance() // ","
	}

	p.advance() // "}"

	return node, nil
}

func parseList(p *Parser) (ListNode, error) {
	var list ListNode

	if p.current.Type != TokenListStart {
		return list, errors.New(`list does not start with "["`)
	}

	p.advance() // "["

	for p.current.Type != TokenEOF && p.current.Type != TokenListEnd {
		lit, err := parseLiteralExpr(p, 0)
		if err != nil {
			return list, fmt.Errorf("failed to parse literal expr in list: %w", err)
		}
		list.Items = append(list.Items, lit)

		if p.current.Type != TokenComma && p.current.Type != TokenListEnd {
			return list, fmt.Errorf(`expected comma or "]" after list item, but got %q`, p.current)
		}

		if p.peek.Type == TokenListEnd {
			return list, errors.New(`extraneous "]" at end of list`)
		}

		if p.current.Type == TokenListEnd {
			break
		}

		p.advance() // ","
	}

	p.advance() // "]"

	return list, nil
}
