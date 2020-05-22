%{
package flatlang
%}

%union {
token int
node *Node
}

%type <node>
Program Var Val CallBody Atom Ident Expr String RawString Literal List ListItem Map MapField

%token <token>
interp comment
ident bool_ int_ float text
',' ':' ';' '(' ')' '[' ']' '{' '}' '='
'\'' '"' '`' '@'


%left <token> '|' '&'
%left <token> '+' '-'
%left <token> '/' '*'
%nonassoc <token> negate '!' '>' '<' lte gte
%nonassoc <token> pipe

/*
p := yylex.(*Parser)
*/

%%

Main : Program { p.Result = $1 };

Program
: { $$ = NewNode(ProgramNode) }
| Program Var ';' { $1.N1($2).T1($3) }
| Program Val ';' { $1.N($2.Nodes...).T1($3) }
;

Ident: ident { $$ = NewNode(IdentNode, $1) };

Var
: Ident '=' Val { $$ = NewNode(VarNode, $2).N1($1).N($3.Nodes...) }
;

Val
: CallBody %prec pipe  { $$ = NewNode(ValNode).N1($1) }
| Val '>' CallBody %prec pipe   { $1.T1($2).N1($3) }
;

CallBody
: Atom { $$ = NewNode(ValNode).N1($1) }
| CallBody Atom     { $1.N1($2) }
;

Atom: Expr %prec pipe | Map | List;

Expr
: Literal
| '(' Expr ')'  { $$ = $2 }

| Expr '/' Expr { $$ = NewOpNode('/', $2).N2($1, $3) }
| Expr '*' Expr { $$ = NewOpNode('*', $2).N2($1, $3) }
| Expr '-' Expr { $$ = NewOpNode('-', $2).N2($1, $3) }
| Expr '+' Expr { $$ = NewOpNode('+', $2).N2($1, $3) }
| Expr '&' Expr { $$ = NewOpNode('&', $2).N2($1, $3) }
| Expr '|' Expr { $$ = NewOpNode('|', $2).N2($1, $3) }

| '<' Expr { $$ = NewOpNode('<', $1).N1($2) }
| lte Expr { $$ = NewOpNode(lte, $1).N1($2) }
| '>' Expr { $$ = NewOpNode('>', $1).N1($2) }
| gte Expr { $$ = NewOpNode(gte, $1).N1($2) }
| '!' Expr { $$ = NewOpNode('!', $1).N1($2) }
| '-' Expr %prec negate { $$ = NewOpNode(negate, $1).N1($2) }
;

Literal
: Ident
| '\'' String '\'' { $$ = $2.T2($1, $3) }
| '"' String '"' { $$ = $2.T2($1, $3) }
| '`' RawString '`' {$$ = $2.T2($1, $3) }
| int_ { $$ = NewNode(IntNode, $1) }
| float { $$ = NewNode(FloatNode, $1) }
| bool_ { $$ = NewNode(BoolNode, $1) }
;

String
: { $$ = NewNode(StringNode) }
| String text { $$ = $1.N1(NewNode(TextNode, $2)) }
;

RawString
: { $$ = NewNode(StringNode) }
| RawString text { $$ = $1.N1(NewNode(TextNode, $2)) }
| RawString interp Expr '}' { $$ = $1.N1(NewNode(InterpNode, $2, $4).N1($3)) }
;

List
: '[' ']' { $$ = NewNode(ListNode, $1, $2) }
| ListItem ']' { $$ = $1.T1($2) }
;

ListItem
: '[' Expr { $$ = NewNode(ListNode, $1).N1($2) }
| ListItem ',' Expr { $1.T1($2).N1($3) }
;

Map
: '{' '}' { $$ = NewNode(MapNode, $1, $2) }
| MapField '}' { $$ = $1.T1($2) }
;

MapField
: '{' Ident ':' Expr { $$ = NewNode(MapNode, $1).N1($2).T1($3).N1($4) }
| MapField ',' Ident ':' Expr { $1.T1($2).N1($3).T1($4).N1($5) }
;