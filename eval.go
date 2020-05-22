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
	res, err := e.Eval(n)
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

func (e *Evaluator) RegisterBuiltin(name string, fn interface{}) error {
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

	if t.IsVariadic() {
		if len(params) < t.NumIn()-1 {
			return fmt.Errorf("%s: expected at least %d param(s), got %d param(s)", name, t.NumIn()-1, len(params))
		}
	} else {
		if len(params) != t.NumIn() {
			return fmt.Errorf("%s: expected exactly %d param(s), got %d param(s)", name, t.NumIn()-1, len(params))
		}
	}

	var pvs []reflect.Value

	if len(params) > 0 {
		pvs = make([]reflect.Value, 0, len(params))

		if t.IsVariadic() {
			i := 0

			for ; i < t.NumIn()-1; i++ {
				pv, it := reflect.ValueOf(params[i]), t.In(i)
				if !pv.Type().AssignableTo(it) {
					return fmt.Errorf("%s: arg %d (%v) is not assignable to %v", name, i, pv.Type(), it)
				}
				pvs = append(pvs, pv)
			}

			vt := t.In(t.NumIn() - 1).Elem()

			for ; i < len(params); i++ {
				pv := reflect.ValueOf(params[i])
				if !pv.Type().AssignableTo(vt) {
					return fmt.Errorf("%s: var arg %d (%v) is not assignable to %v", name, i, pv.Type(), vt)
				}
				pvs = append(pvs, pv)
			}
		} else {
			for i := 0; i < t.NumIn(); i++ {
				pv, it := reflect.ValueOf(params[i]), t.In(i)
				if !pv.Type().AssignableTo(it) {
					return fmt.Errorf("%s: arg %d (%v) is not assignable to %v", name, i, pv.Type(), it)
				}
				pvs = append(pvs, pv)
			}
		}
	}

	out := v.Call(pvs)

	if len(out) == 1 && out[0].Type().Implements(errType) && !out[0].IsNil() {
		return out[0].Interface().(error)
	}

	return nil
}

type methodCall struct {
	name   string
	params []interface{}
}

func (e *Evaluator) Eval(n *Node) (interface{}, error) {
	switch n.Type {
	case ProgramNode:
		results := make([]interface{}, 0, len(n.Nodes))
		for _, node := range n.Nodes {
			res, err := e.Eval(node)
			if err != nil {
				return nil, err
			}

			switch v := res.(type) {
			case methodCall:
				if err := e.dispatch(v.name, v.params...); err != nil {
					return nil, fmt.Errorf("failed to call method %q: %w", v.name, err)
				}
			case []methodCall:
				for _, c := range v {
					if err := e.dispatch(c.name, c.params...); err != nil {
						return nil, fmt.Errorf("failed to call method %q: %w", c.name, err)
					}
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
		rhs := n.Nodes[1:]

		if len(rhs) == 1 {
			res, err := e.Eval(rhs[0])
			if err != nil {
				return nil, fmt.Errorf("failed to eval '%v': %w", sym, err)
			}
			e.sym[sym] = res
			return nil, nil
		}

		calls := make([]methodCall, 0, len(rhs))
		for _, node := range n.Nodes[1:] {
			res, err := e.Eval(node)
			if err != nil {
				return nil, fmt.Errorf("failed to eval '%v': %w", sym, err)
			}

			switch res := res.(type) {
			case []methodCall:
				calls = append(calls, res...)
			default:
				return nil, fmt.Errorf("got unknown type while eval %q's val: %w", sym, err)
			}
		}

		e.sym[sym] = calls

		return nil, nil
	case ValNode:
		if len(n.Nodes) == 1 {
			return e.Eval(n.Nodes[0])
		}

		results := make([]methodCall, 0, len(n.Nodes))
		for i := 0; i < len(n.Nodes); i++ {
			res, err := e.Eval(n.Nodes[i])
			if err != nil {
				return nil, err
			}

			switch res := res.(type) {
			case methodCall:
				results = append(results, res)
			case []methodCall:
				results = append(results, res...)
			default:
				if len(results) == 0 {
					return nil, fmt.Errorf("multiple values may not exist in a single statement unless they serve as parameters for a a method call")
				}
				results[len(results)-1].params = append(results[len(results)-1].params, res)
			}
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
			val, err := e.Eval(node)
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
		val, err := e.Eval(n.Nodes[0])
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
		txt := n.Val(e.lx)

		val, err := strconv.Unquote("\"" + txt + "\"")
		if err != nil {
			return nil, fmt.Errorf("failed to parse string '%v': %w", txt, err)
		}

		return val, nil
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
			val, err := e.Eval(node)
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
			val, err := e.Eval(n.Nodes[i+1])
			if err != nil {
				return nil, err
			}
			vals[ident] = val
		}
		return vals, nil
	case OpNode + negate:
		rhs, err := e.Eval(n.Nodes[0])
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
		lhs, err := e.Eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.Eval(n.Nodes[1])
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
		lhs, err := e.Eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.Eval(n.Nodes[1])
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
		lhs, err := e.Eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.Eval(n.Nodes[1])
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
		lhs, err := e.Eval(n.Nodes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to eval lhs: %w", err)
		}

		rhs, err := e.Eval(n.Nodes[1])
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
