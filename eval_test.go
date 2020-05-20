package flatlang

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

type scope struct {
	vars map[string][]Expr
}

func eval(s *scope, node interface{}) error {
	switch node := node.(type) {
	case Program:
		return evalProgram(s, node)
	case Stmt:
		return evalStmt(s, node)
	case Expr:
		return evalExpr(s, node)
	case Node:
		return evalNode(s, node)
	default:
		panic(node)
	}
}

func evalProgram(s *scope, program Program) error {
	for _, stmt := range program.Stmts {
		if err := evalStmt(s, stmt); err != nil {
			return fmt.Errorf("failed to eval program: %w", err)
		}
	}
	return nil
}

func evalStmt(s *scope, stmt Stmt) error {
	exprs := stmt.Exprs

	switch stmt.Type {
	case StmtTypeAssign:
		for _, expr := range exprs {
			if err := evalExpr(s, expr); err != nil {
				return fmt.Errorf("failed to eval assign expr: %w", err)
			}
		}

		s.vars[stmt.Name] = exprs
	case StmtTypeCall:
		for _, expr := range exprs {
			if err := evalExpr(s, expr); err != nil {
				return fmt.Errorf("failed to eval expr: %w", err)
			}
		}
	}

	return nil
}

func evalExpr(s *scope, expr Expr) error {
	for _, node := range expr.Nodes {
		if err := evalNode(s, node); err != nil {
			return fmt.Errorf("failed to eval (%q): %w", node.Repr(), err)
		}
	}
	return nil
}

func evalNode(s *scope, node Node) error {
	switch node := node.(type) {
	case IdentNode:
		return evalIdent(s, node)
	case StringNode:
		return nil
	default:
		panic(node)
	}
}

func evalIdent(s *scope, ident IdentNode) error {
	if _, exists := s.vars[ident.Val]; exists {
		return nil
	}
	return fmt.Errorf("variable %q is not defined", ident.Val)
}

func TestEvalStmt(t *testing.T) {
	program, err := NewParser(NewLexer(`hello = "world"; world = hello;`)).Parse()
	require.NoError(t, err)

	scope := &scope{vars: make(map[string][]Expr)}
	require.NoError(t, eval(scope, program))

	//spew.Dump(scope)
}
