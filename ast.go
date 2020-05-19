package flatlang

import (
	"fmt"
	"strconv"
	"strings"
)

type Program struct {
	Stmts []Stmt
}

type StmtType int

const (
	StmtTypeAssign StmtType = iota
	StmtTypeCall
)

func (t StmtType) String() string {
	switch t {
	case StmtTypeAssign:
		return "Assign"
	case StmtTypeCall:
		return "Call"
	default:
		return fmt.Sprintf("Unknown(%q)", int(t))
	}
}

type Stmt struct {
	Type StmtType

	Name  string
	Exprs []Expr
}

type Expr struct {
	Nodes []Node
}

func (n Expr) Repr() string { return "" }

type Node interface{ Repr() string }

type UnaryNode struct {
	Type  TokenType
	Right Node
}

func (n UnaryNode) Repr() string {
	var b strings.Builder
	b.WriteByte('(')
	b.WriteString(n.Type.String())
	b.WriteString(n.Right.Repr())
	b.WriteByte(')')
	return b.String()
}

type BinaryNode struct {
	Left  Node
	Type  TokenType
	Right Node
}

func (n BinaryNode) Repr() string {
	var b strings.Builder
	b.WriteByte('(')
	b.WriteString(n.Left.Repr())
	b.WriteByte(' ')
	b.WriteString(n.Type.String())
	b.WriteByte(' ')
	b.WriteString(n.Right.Repr())
	b.WriteByte(')')
	return b.String()
}

type IdentNode struct{ Val string }

func (n IdentNode) Repr() string { return n.Val }

type BoolNode struct{ Val bool }

func (n BoolNode) Repr() string {
	if n.Val {
		return "true"
	} else {
		return "false"
	}
}

type StringNode struct{ Val string }

func (n StringNode) Repr() string { return strconv.Quote(n.Val) }

type IntNode struct{ Val int64 }

func (n IntNode) Repr() string { return strconv.FormatInt(n.Val, 10) }

type FloatNode struct{ Val float64 }

func (n FloatNode) Repr() string { return strconv.FormatFloat(n.Val, 'f', -1, 64) }

type OpNode struct{ Op string }

type SetOpNode struct{ Op string }

type CmpNode struct{ Op string }

type ListNode struct {
	Items []Node
}

func (n ListNode) Repr() string { return "" }

type MapNode struct {
	Fields []Field
}

func (n MapNode) Repr() string { return "" }

type Field struct {
	Key string
	Val Node
}
