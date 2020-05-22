package flatlang

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

type Evaluator struct {
	lx  *Lexer
	sym map[string]interface{}
}

func newEval(lx *Lexer) *Evaluator {
	return &Evaluator{lx: lx, sym: make(map[string]interface{})}
}

func (e *Evaluator) eval(n *Node) (interface{}, error) {
	switch n.Type {
	case ProgramNode:
		results := make([]interface{}, 0, len(n.Nodes))
		for _, node := range n.Nodes {
			res, err := e.eval(node)
			if err != nil {
				return nil, err
			}
			results = append(results, res)
		}
		return results, nil
	case IdentNode:
		sym := n.Val(e.lx)
		val, recorded := e.sym[sym]
		if recorded {
			return val, nil
		}
		return nil, fmt.Errorf("unknown symbol '%v'", sym)
	case VarNode:
		lhs := n.Nodes[0].Val(e.lx)
		rhs, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval '%v'", lhs)
		}
		e.sym[lhs] = rhs
		return nil, nil
	case ValNode:
		if len(n.Nodes) == 1 {
			return e.eval(n.Nodes[0])
		}

		results := make([]interface{}, 0, len(n.Nodes))
		for i := 0; i < len(n.Nodes); i++ {
			res, err := e.eval(n.Nodes[i])
			if err != nil {
				return nil, err
			}
			results = append(results, res)
		}
		return results, nil
	case BoolNode:
		val := n.Val(e.lx)
		switch val {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
		return nil, fmt.Errorf("got malformed bool %q", val)
	case StringNode:
		res := ""

		for _, node := range n.Nodes {
			val, err := e.eval(node)
			if err != nil {
				return nil, err
			}
			txt, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("got %q while evaluating string", val)
			}
			res += txt
		}
		return res, nil
	case InterpNode:
		val, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, err
		}
		switch val := val.(type) {
		case string:
			return val, nil
		case int64:
			return strconv.FormatInt(val, 10), nil
		case float64:
			return strconv.FormatFloat(val, 'g', -1, 64), nil
		case bool:
			if val {
				return "true", nil
			}
			return "false", nil
		}
		return nil, fmt.Errorf("unable to interpolate '%v' into a string", val)
	case TextNode:
		return n.Val(e.lx), nil
	case IntNode:
		val, err := strconv.ParseInt(n.Val(e.lx), 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to eval int: %w", err)
		}
		return val, nil
	case FloatNode:
		val, err := strconv.ParseFloat(n.Val(e.lx), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to eval float: %w", err)
		}
		return val, nil
	case ListNode:
		vals := make([]interface{}, 0, len(n.Nodes))
		for _, node := range n.Nodes {
			val, err := e.eval(node)
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		}
		return vals, nil
	case MapNode:
		vals := make(map[string]interface{}, len(n.Nodes)/2)
		for i := 0; i < len(n.Nodes); i += 2 {
			ident := n.Nodes[i].Val(e.lx)
			val, err := e.eval(n.Nodes[i+1])
			if err != nil {
				return nil, err
			}
			vals[ident] = val
		}
		return vals, nil
	case OpNode + negate:
		rhs, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch rhs := rhs.(type) {
		case int64:
			return -rhs, nil
		case float64:
			return -rhs, nil
		}

		return nil, fmt.Errorf("unable to negate '%v'", n.Type)
	case OpNode + '+':
		lhs, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch lhs := lhs.(type) {
		case []interface{}:
			switch rhs := rhs.(type) {
			case []interface{}:
				return append(lhs, rhs...), nil
			}
		case string:
			switch r := rhs.(type) {
			case string:
				return lhs + r, nil
			}
		case int64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs + rhs, nil
			case float64:
				return float64(lhs) + rhs, nil
			}
		case float64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs + float64(rhs), nil
			case float64:
				return lhs + rhs, nil
			}
		}

		return nil, fmt.Errorf("cannot eval '%v' + '%v'", lhs, rhs)
	case OpNode + '-':
		lhs, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch lhs := lhs.(type) {
		case int64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs - rhs, nil
			case float64:
				return float64(lhs) - rhs, nil
			}
		case float64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs - float64(rhs), nil
			case float64:
				return lhs - rhs, nil
			}
		}

		return nil, fmt.Errorf("cannot eval '%v' - '%v'", lhs, rhs)
	case OpNode + '*':
		lhs, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch lhs := lhs.(type) {
		case int64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs * rhs, nil
			case float64:
				return float64(lhs) * rhs, nil
			}
		case float64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs * float64(rhs), nil
			case float64:
				return lhs * rhs, nil
			}
		}
		return nil, fmt.Errorf("cannot eval '%v' * '%v'", lhs, rhs)
	case OpNode + '/':
		lhs, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch lhs := lhs.(type) {
		case int64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs / rhs, nil
			case float64:
				return float64(lhs) / rhs, nil
			}
		case float64:
			switch rhs := rhs.(type) {
			case int64:
				return lhs / float64(rhs), nil
			case float64:
				return lhs / rhs, nil
			}
		}
		return nil, fmt.Errorf("cannot eval '%v' / '%v'", lhs, rhs)
	}

	spew.Dump(n)
	panic(fmt.Sprintf("unknown node type '%v(%d)'", n.Type, n.Type))
}

func Eval(lx *Lexer, n *Node) (interface{}, error) {
	e := newEval(lx)
	res, err := e.eval(n)
	if err != nil {
		return nil, err
	}
	spew.Dump(e.sym)
	return res, nil
}

func TestEval(t *testing.T) {
	src := []byte("hi = 123 + 4.0; there = `this is a ${hi + 5} test`;")

	lx, err := Lex(src, "")
	require.NoError(t, err)

	px, err := Parse(lx)
	require.NoError(t, err)

	fmt.Printf("Evaluating %q.\n\n", src[:len(src)-1])

	res, err := Eval(lx, px.Result)
	require.NoError(t, err)

	fmt.Println()

	spew.Dump(res)
}
