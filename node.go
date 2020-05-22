package flatlang

type NodeType int

const (
	ProgramNode NodeType = iota
	VarNode
	ValNode
	ExprNode
	InterpNode
	IdentNode
	BoolNode
	IntNode
	FloatNode
	StringNode
	TextNode
	ListNode
	MapNode

	OpNode
)

var NodeString = [...]string{
	ProgramNode: "program",
	VarNode:     "var",
	ValNode:     "val",
	ExprNode:    "expr",
	InterpNode:  "interp",
	IdentNode:   "ident",
	BoolNode:    "bool",
	IntNode:     "int",
	FloatNode:   "float",
	StringNode:  "string",
	TextNode:    "text",
	ListNode:    "list",
	MapNode:     "map",

	OpNode + negate: "-",
	OpNode + '+':    "+",
	OpNode + '-':    "-",
	OpNode + '*':    "*",
	OpNode + '/':    "/",
	OpNode + '>':    ">",
	OpNode + '<':    "<",
	OpNode + gte:    ">=",
	OpNode + lte:    "<=",
	OpNode + '!':    "!",
	OpNode + '&':    "&",
	OpNode + '|':    "|",
}

func (t NodeType) String() string { return NodeString[t] }

type Node struct {
	Type   NodeType
	Tokens []int
	Nodes  []*Node
}

func NewNode(t NodeType, tokens ...int) *Node   { return &Node{Type: t, Tokens: tokens} }
func NewOpNode(t NodeType, tokens ...int) *Node { return NewNode(OpNode+t, tokens...) }

func (n *Node) T(tokens ...int) *Node   { n.Tokens = append(n.Tokens, tokens...); return n }
func (n *Node) T1(t0 int) *Node         { n.Tokens = append(n.Tokens, t0); return n }
func (n *Node) T2(t0, t1 int) *Node     { n.Tokens = append(n.Tokens, t0, t1); return n }
func (n *Node) T3(t0, t1, t2 int) *Node { n.Tokens = append(n.Tokens, t0, t1, t2); return n }

func (n *Node) N(nodes ...*Node) *Node    { n.Nodes = append(n.Nodes, nodes...); return n }
func (n *Node) N1(n0 *Node) *Node         { n.Nodes = append(n.Nodes, n0); return n }
func (n *Node) N2(n0, n1 *Node) *Node     { n.Nodes = append(n.Nodes, n0, n1); return n }
func (n *Node) N3(n0, n1, n2 *Node) *Node { n.Nodes = append(n.Nodes, n0, n1, n2); return n }

func (n Node) Val(lx *Lexer) string {
	return string(lx.Data[lx.Tokens[n.Tokens[0]].Pos:lx.Tokens[n.Tokens[0]].End])
}

func (n Node) Format(lx *Lexer) string {
	buf := n.Type.String()
	if buf == "" {
		panic("unknown node type")
	}
	for _, child := range n.Nodes {
		buf += " " + child.Format(lx)
	}

	if len(n.Tokens) > 0 {
		repr := n.Val(lx)

		switch n.Type {
		case IdentNode:
			buf += " |" + repr + "|"
		case BoolNode:
			buf += " " + repr
		case FloatNode, IntNode:
			buf += " " + repr
		case TextNode:
			buf += ` "` + repr + `"`
		}
	}

	return "(" + buf + ")"
}
