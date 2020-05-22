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
	case ValNode:
		results := make([]interface{}, 0, len(n.Nodes))
		for i := 0; i < len(n.Nodes); i++ {
			res, err := e.eval(n.Nodes[i])
			if err != nil {
				return nil, err
			}
			results = append(results, res)

			//if i != 0 {
			//	continue
			//}
			//
			//if len(results) == 1 {
			//	continue
			//}
			//
			//if _, ok := results[0].(func()); !ok {
			//	return nil, fmt.Errorf("first ")
			//}
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
	case OpNode + negate:
		r, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch val := r.(type) {
		case int64:
			return -val, nil
		case float64:
			return -val, nil
		}

		return nil, fmt.Errorf("unable to negate '%v'", n.Type)
	case OpNode + '+':
		l, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		r, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch l := l.(type) {
		case string:
			switch r := r.(type) {
			case string:
				return l + r, nil
			}
		case int64:
			switch r := r.(type) {
			case int64:
				return l + r, nil
			case float64:
				return float64(l) + r, nil
			}
		case float64:
			switch r := r.(type) {
			case int64:
				return l + float64(r), nil
			case float64:
				return l + r, nil
			}
		}

		return nil, fmt.Errorf("cannot eval '%v' + '%v'", l, r)
	case OpNode + '-':
		l, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		r, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch l := l.(type) {
		case int64:
			switch r := r.(type) {
			case int64:
				return l - r, nil
			case float64:
				return float64(l) - r, nil
			}
		case float64:
			switch r := r.(type) {
			case int64:
				return l - float64(r), nil
			case float64:
				return l - r, nil
			}
		}

		return nil, fmt.Errorf("cannot eval '%v' - '%v'", l, r)
	case OpNode + '*':
		l, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		r, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch l := l.(type) {
		case int64:
			switch r := r.(type) {
			case int64:
				return l * r, nil
			case float64:
				return float64(l) * r, nil
			}
		case float64:
			switch r := r.(type) {
			case int64:
				return l * float64(r), nil
			case float64:
				return l * r, nil
			}
		}
		return nil, fmt.Errorf("cannot eval '%v' * '%v'", l, r)
	case OpNode + '/':
		l, err := e.eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		r, err := e.eval(n.Nodes[1])
		if err != nil {
			return nil, fmt.Errorf("failed to eval rhs: %w", err)
		}

		switch l := l.(type) {
		case int64:
			switch r := r.(type) {
			case int64:
				return l / r, nil
			case float64:
				return float64(l) / r, nil
			}
		case float64:
			switch r := r.(type) {
			case int64:
				return l / float64(r), nil
			case float64:
				return l / r, nil
			}
		}
		return nil, fmt.Errorf("cannot eval '%v' / '%v'", l, r)
	}

	panic(fmt.Sprintf("unknown node type '%v'", n.Type))
}

func Eval(lx *Lexer, n *Node) (interface{}, error) {
	return newEval(lx).eval(n)
}

func TestEval(t *testing.T) {
	src := []byte(`432 + (291 * (26 - 32.0));`)

	lx, err := Lex(src, "")
	require.NoError(t, err)

	px, err := Parse(lx)
	require.NoError(t, err)

	res, err := Eval(lx, px.Result)
	require.NoError(t, err)

	fmt.Printf("Evaluating %q.\n\n", src[:len(src)-1])
	spew.Dump(res)
}
