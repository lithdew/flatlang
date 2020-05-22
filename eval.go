package flatlang

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"strconv"
)

type Evaluator struct {
	lx       *Lexer
	sym      map[string]interface{}
	builtins map[string]reflect.Value
}

func Eval(lx *Lexer, n *Node) (interface{}, error) {
	e := NewEval(lx)
	res, err := e.eval(n)
	if err != nil {
		return nil, err
	}
	if len(e.sym) > 0 {
		spew.Dump(e.sym)
	}
	return res, nil
}

func NewEval(lx *Lexer) *Evaluator {
	return &Evaluator{
		lx:       lx,
		sym:      make(map[string]interface{}),
		builtins: make(map[string]reflect.Value),
	}
}

var errType = reflect.TypeOf((*error)(nil)).Elem()

func (e *Evaluator) register(name string, fn interface{}) error {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("%q is not a func", name)
	}

	t := v.Type()

	if t.NumOut() > 1 {
		return fmt.Errorf("methods may only return an error at most")
	}
	if t.NumOut() == 1 && !t.Out(0).Implements(errType) {
		return fmt.Errorf("return val of method is expected to be an error, but got %v", t.Out(0))
	}

	e.builtins[name] = v
	return nil
}

func (e *Evaluator) dispatch(name string, params ...interface{}) error {
	v, exists := e.builtins[name]
	if !exists {
		return fmt.Errorf("method %q not registered", name)
	}

	t := v.Type()

	if !t.IsVariadic() && t.NumIn() != len(params) {
		return fmt.Errorf("%s: expected %d params, got %d params", name, t.NumIn(), len(params))
	}

	var pvs []reflect.Value

	if len(params) > 0 {
		pvs = make([]reflect.Value, 0, len(params))

		for i := 0; i < len(params); i++ {
			idx := i
			if idx >= t.NumIn() {
				idx = t.NumIn() - 1
			}

			pv, iv := reflect.ValueOf(params[i]), t.In(idx)
			if pt := pv.Type(); !t.IsVariadic() && !pt.AssignableTo(iv) {
				return fmt.Errorf("%s: param %d (%v) is not assignable to in type %v", name, i, pt, iv)
			}
			pvs = append(pvs, pv)
		}
	}

	out := v.Call(pvs)

	if len(out) == 1 && out[0].Type().Implements(errType) {
		return out[0].Interface().(error)
	}

	return nil
}

type methodCall struct {
	name   string
	params []interface{}
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

			switch val := res.(type) {
			case methodCall:
				if err := e.dispatch(val.name, val.params...); err != nil {
					return nil, fmt.Errorf("failed to call method %q: %w", val.name, err)
				}
				results = append(results, nil)
			default:
				results = append(results, res)
			}
		}
		return results, nil
	case IdentNode:
		sym := n.Val(e.lx)
		if val, recorded := e.sym[sym]; recorded {
			return val, nil
		}
		if _, exists := e.builtins[sym]; exists {
			return methodCall{name: sym}, nil
		}
		return nil, fmt.Errorf("unknown symbol '%v'", sym)
	case VarNode:
		sym := n.Nodes[0].Val(e.lx)

		results := make([]interface{}, 0, len(n.Nodes[1:]))
		for _, node := range n.Nodes[1:] {
			res, err := e.eval(node)
			if err != nil {
				return nil, fmt.Errorf("failed to eval '%v': %w", sym, err)
			}
			results = append(results, res)
		}

		if len(results) == 1 {
			e.sym[sym] = results[0]
		} else {
			e.sym[sym] = results
		}
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
			switch res := res.(type) {
			case []interface{}:
				results = append(results, res...)
			default:
				results = append(results, res)
			}
		}
		if call, ok := results[0].(methodCall); ok {
			if len(results) > 1 {
				call.params = results[1:]
			}
			return call, nil
		}
		return nil, fmt.Errorf("multiple values may not exist in a single statement unless it is a method call: got %v", results)
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
