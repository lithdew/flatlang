package flatlang

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func N(t testing.TB, symbol string) Node {
	t.Helper()

	node, err := parseLiteralExpr(NewParser(NewLexer(symbol)), 0)
	require.NoError(t, err)
	return node
}

func LN(t testing.TB, src string) ListNode {
	t.Helper()

	list, err := parseList(NewParser(NewLexer(src)))
	require.NoError(t, err)
	return list
}

func MN(t testing.TB, src string) MapNode {
	t.Helper()

	node, err := parseMap(NewParser(NewLexer(src)))
	require.NoError(t, err)
	return node
}

func MF(t testing.TB, key, val string) Field {
	t.Helper()

	node, err := parseLiteralExpr(NewParser(NewLexer(val)), 0)
	require.NoError(t, err)

	return Field{Key: key, Val: node}
}

func TestParse(t *testing.T) {
	src, err := ioutil.ReadFile("testdata/test.fbs")
	require.NoError(t, err)

	program, err := NewParser(NewLexer(string(src))).Parse()
	require.NoError(t, err)

	//spew.Dump(program)
	_ = program
}

func TestParseStmt(t *testing.T) {
	cases := []struct {
		src      string
		expected Stmt
	}{
		{
			src: `handler = what "hello"|(world + ("what" + "test")) world`,
			expected: Stmt{
				Type: StmtTypeAssign,
				Name: "handler",
				Exprs: []Expr{
					{
						Nodes: []Node{
							N(t, `what`),
							N(t, `"hello"|(world + ("what" + "test"))`),
							N(t, `world`),
						},
					},
				},
			},
		},
		{
			src: `handler = get "/item/:id" > require offset >= 100 & < 1000 > set "hello"|>=3`,
			expected: Stmt{
				Type: StmtTypeAssign,
				Name: "handler",
				Exprs: []Expr{
					{
						Nodes: []Node{
							N(t, `get`),
							N(t, "`/item/:id`"),
						},
					},
					{
						Nodes: []Node{
							N(t, `require`),
							N(t, `offset`),
							N(t, `>= 100 & < 1000`),
						},
					},
					{
						Nodes: []Node{
							N(t, `set`),
							N(t, `"hello"|>=3`),
						},
					},
				},
			},
		},
		{
			src: `helper = > require true > set {hello: false};`,
			expected: Stmt{
				Type: StmtTypeAssign,
				Name: "helper",
				Exprs: []Expr{
					{
						Nodes: []Node{
							N(t, `require`),
							N(t, `true`),
						},
					},
					{
						Nodes: []Node{
							N(t, `set`),
							MN(t, `{hello: false}`),
						},
					},
				},
			},
		},
		{
			src: `get "/item/:id" > require offset >= 100 & < 1000 > set "hello"|>=3`,
			expected: Stmt{
				Type: StmtTypeCall,
				Exprs: []Expr{
					{
						Nodes: []Node{
							N(t, `get`),
							N(t, "`/item/:id`"),
						},
					},
					{
						Nodes: []Node{
							N(t, `require`),
							N(t, `offset`),
							N(t, `>= 100 & < 1000`),
						},
					},
					{
						Nodes: []Node{
							N(t, `set`),
							N(t, `"hello"|>=3`),
						},
					},
				},
			},
		},
		{src: `get "/item/:id" > "hello" 1 2 3`},
	}

	for _, test := range cases {
		actual, err := parseStmt(NewParser(NewLexer(test.src)))
		if test.expected.Exprs != nil {
			require.NoError(t, err)
			require.EqualValues(t, test.expected, actual)
			continue
		}
		require.Error(t, err)
	}
}

func TestParseExprs(t *testing.T) {
	cases := []struct {
		src      string
		expected []Expr
	}{
		{
			src: `get "/item/:id" > require offset >= 100 & < 1000 > set "hello"|>=3`,
			expected: []Expr{
				{
					Nodes: []Node{
						N(t, `get`),
						N(t, "`/item/:id`"),
					},
				},
				{
					Nodes: []Node{
						N(t, `require`),
						N(t, `offset`),
						N(t, `>= 100 & < 1000`),
					},
				},
				{
					Nodes: []Node{
						N(t, `set`),
						N(t, `"hello"|>=3`),
					},
				},
			},
		},
		{
			src: `hello [1, 2, 3] > a ["test"] true`,
			expected: []Expr{
				{
					Nodes: []Node{
						N(t, `hello`),
						LN(t, `[1, 2, 3]`),
					},
				},
				{
					Nodes: []Node{
						N(t, `a`),
						LN(t, `["test"]`),
						N(t, `true`),
					},
				},
			},
		},
		{
			src: `this {a: >=1, b: 4} > works {c: false, d: 30.0} false`,
			expected: []Expr{
				{
					Nodes: []Node{
						N(t, `this`),
						MN(t, `{a: >=1, b: 4}`),
					},
				},
				{
					Nodes: []Node{
						N(t, `works`),
						MN(t, `{c: false, d: 30.0}`),
						N(t, `false`),
					},
				},
			},
		},
		{
			src: `the [1, 2, "test"] > works {c: false, d: 30.0} false; a b c`,
			expected: []Expr{
				{
					Nodes: []Node{
						N(t, `the`),
						LN(t, `[1, 2, "test"]`),
					},
				},
				{
					Nodes: []Node{
						N(t, `works`),
						MN(t, `{c: false, d: 30.0}`),
						N(t, `false`),
					},
				},
			},
		},
	}

	for _, test := range cases {
		p := NewParser(NewLexer(test.src))

		for _, expected := range test.expected {
			actual, err := parseExpr(p)
			require.NoError(t, err)
			require.EqualValues(t, expected, actual)
		}

		p = NewParser(NewLexer(test.src))

		actual, err := parseExprs(p)
		require.NoError(t, err)
		require.EqualValues(t, test.expected, actual)
	}
}

func TestParseList(t *testing.T) {
	cases := []struct {
		src      string
		expected *ListNode
	}{
		{
			src: `[1, 2, 3, "hello", 4.3]`,
			expected: &ListNode{
				Items: []Node{
					N(t, `1`),
					N(t, `2`),
					N(t, `3`),
					N(t, `"hello"`),
					N(t, `4.3`),
				},
			},
		},
		{
			src: `[1, 2|"hello"&<=3, false]`,
			expected: &ListNode{
				Items: []Node{
					N(t, `1`),
					N(t, `2|"hello"&<=3`),
					N(t, `false`),
				},
			},
		},
		{src: `[]`, expected: &ListNode{}},
		{src: `[1, 2, false,]`},
		{src: `[, 2, false]`},
		{src: `[4, 2, false`},
		{src: `[1, 2, 3, true,]`},
	}

	for _, test := range cases {
		actual, err := parseList(NewParser(NewLexer(test.src)))
		if test.expected != nil {
			require.NoError(t, err)
			require.EqualValues(t, *test.expected, actual)
			continue
		}
		require.Error(t, err)
	}
}

func TestParseMap(t *testing.T) {
	cases := []struct {
		src      string
		expected *MapNode
	}{
		{
			src: `{hello: "world", test: >1 & <100}`,
			expected: &MapNode{
				Fields: []Field{
					MF(t, "hello", `"world"`),
					MF(t, "test", `>1 & <100`),
				},
			},
		},
		{
			src: `{a: b, b: 1|"test", c: false}`,
			expected: &MapNode{
				Fields: []Field{
					MF(t, "a", "b"),
					MF(t, "b", `1|"test"`),
					MF(t, "c", "false"),
				},
			},
		},
		{src: `{}`, expected: &MapNode{}},
		{src: `{hello:`},
		{src: `{"hello": "world"}`},
		{src: `{hello: "world"`},
		{src: `{hello: "world",}`},
		{src: `{hello: }`},
		{src: `{hello:,}`},
		{src: `{hello:,,,}`},
		{src: `{hello: {test: 1}, test: false}`},
		{src: `{hello: [test, "a", 'b'], test: true}`},
	}

	for _, test := range cases {
		actual, err := parseMap(NewParser(NewLexer(test.src)))
		if test.expected != nil {
			require.NoError(t, err)
			require.EqualValues(t, *test.expected, actual)
			continue
		}
		require.Error(t, err)
	}
}
