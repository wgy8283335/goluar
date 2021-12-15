/*
	See Copyright Notice at LICENSE file
*/

package compiler

/*
	Describe syntax in EBNF:

	exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef | prefixexp | tableconstructor | exp binop exp | unop exp
	prefixexp ::= var | functioncall | ‘(’ exp ‘)’
	var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
	functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
*/
type Exp interface{}

/*
	nil
*/
type NilExp struct {
	Line int // line number
}

/*
	true
*/
type TrueExp struct {
	Line int // line number
}

/*
	false
*/
type FalseExp struct {
	Line int // line number
}

/*
	...
*/
type VarargExp struct {
	Line int // line number
}

/*
	Numeral
*/
type IntegerExp struct {
	Line int   // line number
	Val  int64 //value
}

/*
	Numeral
*/
type FloatExp struct {
	Line int     // line number
	Val  float64 //value
}

/*
	LiteralString
*/
type StringExp struct {
	Line int    // line number
	Str  string //value
}

/*
	unop exp
*/
type UnopExp struct {
	Line int // line number
	Op   int // operator
	Exp  Exp
}

/*
	exp binop exp
*/
type BinopExp struct {
	Line int // line number
	Op   int // operator
	Exp1 Exp // expression
	Exp2 Exp // expression
}

/*
	Concat operator, like: ..
*/
type ConcatExp struct { //此处有待考虑，EBNF描述中是否存在??
	Line int   // last line number
	Exps []Exp // expression to be concated
}

/*
	tableconstructor ::= ‘{’ [fieldlist] ‘}’
	fieldlist ::= field {fieldsep field} [fieldsep]
	field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
	fieldsep ::= ‘,’ | ‘;’
*/
type TableConstructorExp struct {
	Line     int   // line number of `{`
	LastLine int   // line number of `}`
	KeyExps  []Exp // expression in key of table
	ValExps  []Exp // expression in value of table
}

/*
	functiondef ::= function funcbody
	funcbody ::= ‘(’ [parlist] ‘)’ block end
	parlist ::= namelist [‘,’ ‘...’] | ‘...’
	namelist ::= Name {‘,’ Name}
*/
type FuncDefExp struct {
	Line     int      // line number at the begin of expression
	LastLine int      // line number at the end of expression
	ParList  []string // parameter list in ()
	IsVararg bool     // whether is variable argument
	Block    *Block   // block in function body
}

/*
	prefixexp ::= var | funcioncall	| ‘(’ exp ‘)’
	var ::= Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
	funcioncall ::= prefixexp args | prefixexp ':' Name args

	All of these are split into four kinds: nameexp, parensexp, tableexp, funcioncall.
*/

/*
	nameexp ::= Name
*/
type NameExp struct {
	Line int    // line number
	Name string // expression name
}

/*
	parensexp ::= '(' exp ')'
*/
type ParensExp struct {
	Exp Exp //expresion in ()
}

/*
	tableexp ::= prefixexp ‘[’ exp ‘]’
				| prefixexp ‘.’ Name
*/
type TableAccessExp struct {
	LastLine  int // line number of `]`
	PrefixExp Exp // prefix expression 'a.k', 'a.k' is the same as a["k"].
	KeyExp    Exp // expression in key of table
}

/*
	funcioncall ::= prefixexp args | prefixexp ':' Name args
*/
type FuncCallExp struct {
	Line      int        // line number of `(`
	LastLine  int        // last line number of ')'
	PrefixExp Exp        // prefix expression
	NameExp   *StringExp // function name
	Args      []Exp      // arguments in function
}
