/*
	See Copyright Notice at LICENSE file.
*/
package compiler

import . "goluar/common"

/*
	@description
		Describe syntax in EBNF:
		explist ::= exp {‘,’ exp}
	@param
		lexer	Lexer	"Lexical analyzer"
	@result
		exps	[]Exp		"Exp is defined in ast_exp.go"
*/
func parseExpList(lexer *Lexer) []Exp {
	exps := make([]Exp, 0, 4)
	exps = append(exps, parseExp(lexer))
	for lexer.LookAhead() == LEX_SEP_COMMA {
		lexer.NextToken()
		exps = append(exps, parseExp(lexer))
	}
	return exps
}

/*
	@description
		Describe syntax in EBNF:
		exp ::= nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
				prefixexp | tableconstructor | exp binop exp | unop exp
		exp ::= ExpCompare {(or|and) ExpCompare}
		ExpCompare ::= expBitOr {(‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’) expBitOr}
		expBitOp ::= expConcat {('|' |  '~' | '&' | 'shift') expConcat}
		expConcat ::= expOpMath {‘..’ expOpMath}
		expOpMath ::= expUniOp {(‘+’ | ‘-’ | ‘*’ | ‘/’ | ‘//’ | ‘%’) expUniOp}
		expUniOp ::= {(‘not’ | ‘#’ | ‘-’ | ‘~’)} expPow
		expPow ::= expOther {‘^’ expUniOp}
		expOther ::= nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
					prefixexp | tableconstructor
	@param
		lexer	Lexer	"Lexical analyzer"
	@result
		exp	Exp		"Exp is defined in ast_exp.go"

*/
func parseExp(lexer *Lexer) Exp {
	exp := parseExpCompare(lexer)
	for {
		if lexer.LookAhead() == LEX_OP_AND || lexer.LookAhead() == LEX_OP_OR {
			line, op, _ := lexer.NextToken()
			landor := &BinopExp{line, op, exp, parseExpCompare(lexer)}
			exp = optimizeLogicalAndOr(landor)
		} else {
			return exp
		}
	}
	return exp
}

/*
	@description
		Describe syntax in EBNF:
		prefixexp ::= var | functioncall | ‘(’ exp ‘)’
		var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
		functioncall ::=  prefixexp args | prefixexp ‘:’ Name args

		After simplify the EBNF of prefixexp：
		prefixexp ::= Name
			| ‘(’ exp ‘)’
			| prefixexp ‘[’ exp ‘]’
			| prefixexp ‘.’ Name
			| prefixexp [‘:’ Name] args
	@param
		lexer	Lexer	"Lexical analyzer"
	@result
		exp	Exp		"Exp is defined in ast_exp.go"
*/
func parsePrefixExp(lexer *Lexer) Exp {
	var exp Exp
	if lexer.LookAhead() == LEX_IDENTIFIER {
		line, name := lexer.NextIdentifier() // Name
		exp = &NameExp{line, name}
	} else { // ‘(’ exp ‘)’
		exp = parseParensExp(lexer)
	}
	return finishPrefixExp(lexer, exp)
}

/*
	@description
		Describe syntax in EBNF:
		functiondef ::= function funcbody
		funcbody ::= ‘(’ [parlist] ‘)’ block end
	@param
		lexer	Lexer	"Lexical analyzer"
	@result
		funcDefExp	FuncDefExp	"FuncDefExp is defined in ast_exp.go"
*/
func parseFuncDefExp(lexer *Lexer) *FuncDefExp {
	line := lexer.Line()                             // function
	lexer.NextTokenOfKind(LEX_SEP_LPAREN)            // (
	parList, isVararg := parseParList(lexer)         // [parlist]
	lexer.NextTokenOfKind(LEX_SEP_RPAREN)            // )
	block := parseBlock(lexer)                       // block
	lastLine, _ := lexer.NextTokenOfKind(LEX_KW_END) // end
	return &FuncDefExp{line, lastLine, parList, isVararg, block}
}

// ‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’
func parseExpCompare(lexer *Lexer) Exp {
	exp := parseExpConcat(lexer)
	for {
		if lexer.LookAhead() == LEX_OP_LT ||
			lexer.LookAhead() == LEX_OP_GT ||
			lexer.LookAhead() == LEX_OP_NE ||
			lexer.LookAhead() == LEX_OP_LE ||
			lexer.LookAhead() == LEX_OP_GE ||
			lexer.LookAhead() == LEX_OP_EQ {
			line, op, _ := lexer.NextToken()
			exp = &BinopExp{line, op, exp, parseExpConcat(lexer)}
		} else {
			return exp
		}
	}
	return exp
}

// '..'
func parseExpConcat(lexer *Lexer) Exp {
	exp := parseExpOpMath(lexer)
	if lexer.LookAhead() != LEX_OP_CONCAT {
		return exp
	}

	line := 0
	exps := []Exp{exp}
	for lexer.LookAhead() == LEX_OP_CONCAT {
		line, _, _ = lexer.NextToken()
		exps = append(exps, parseExpOpMath(lexer))
	}
	return &ConcatExp{line, exps}
}

// '+' | '-' | '*' | '%' | '/'
func parseExpOpMath(lexer *Lexer) Exp {
	exp := parseExpUniOp(lexer)
	for {
		switch lexer.LookAhead() {
		case LEX_OP_MUL, LEX_OP_MOD, LEX_OP_DIV,
			LEX_OP_ADD, LEX_OP_SUB:
			line, op, _ := lexer.NextToken()
			arith := &BinopExp{line, op, exp, parseExpUniOp(lexer)}
			exp = optimizeArithBinaryOp(arith)
		default:
			return exp
		}
	}
	return exp
}

// '-' | '#' | 'not'
func parseExpUniOp(lexer *Lexer) Exp {
	switch lexer.LookAhead() {
	case LEX_OP_UNM, LEX_OP_LEN, LEX_OP_NOT:
		line, op, _ := lexer.NextToken()
		exp := &UnopExp{line, op, parseExpUniOp(lexer)}
		return optimizeUnaryOp(exp)
	}
	return parseExpPow(lexer)
}

// '^'
func parseExpPow(lexer *Lexer) Exp { // pow is right associative
	exp := parseExpOther(lexer)
	if lexer.LookAhead() == LEX_OP_POW {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{line, op, exp, parseExpUniOp(lexer)}
	}
	return optimizePow(exp)
}

//  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef | prefixexp | tableconstructor
func parseExpOther(lexer *Lexer) Exp {
	switch lexer.LookAhead() {
	case LEX_VARARG: // ...
		line, _, _ := lexer.NextToken()
		return &VarargExp{line}
	case LEX_KW_NIL: // nil
		line, _, _ := lexer.NextToken()
		return &NilExp{line}
	case LEX_KW_TRUE: // true
		line, _, _ := lexer.NextToken()
		return &TrueExp{line}
	case LEX_KW_FALSE: // false
		line, _, _ := lexer.NextToken()
		return &FalseExp{line}
	case LEX_STRING: // LiteralString
		line, _, token := lexer.NextToken()
		return &StringExp{line, token}
	case LEX_NUMBER: // Numeral
		return parseNumberExp(lexer)
	case LEX_SEP_LCURLY: // tableconstructor
		return parseTableConstructorExp(lexer)
	case LEX_KW_FUNCTION: // functiondef
		lexer.NextToken()
		return parseFuncDefExp(lexer)
	default: // prefixexp
		return parsePrefixExp(lexer)
	}
}

// Numeral
func parseNumberExp(lexer *Lexer) Exp {
	line, _, token := lexer.NextToken()
	if i, ok := ParseInteger(token); ok {
		return &IntegerExp{line, i}
	} else if f, ok := ParseFloat(token); ok {
		return &FloatExp{line, f}
	} else { // todo
		panic("not a number: " + token)
	}
}

// [parlist]
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
func parseParList(lexer *Lexer) (names []string, isVararg bool) {
	switch lexer.LookAhead() {
	case LEX_SEP_RPAREN:
		return nil, false
	case LEX_VARARG:
		lexer.NextToken()
		return nil, true
	}
	_, name := lexer.NextIdentifier()
	names = append(names, name)
	for lexer.LookAhead() == LEX_SEP_COMMA {
		lexer.NextToken()
		if lexer.LookAhead() == LEX_IDENTIFIER {
			_, name := lexer.NextIdentifier()
			names = append(names, name)
		} else {
			lexer.NextTokenOfKind(LEX_VARARG)
			isVararg = true
			break
		}
	}
	return
}

// tableconstructor ::= ‘{’ [fieldlist] ‘}’
func parseTableConstructorExp(lexer *Lexer) *TableConstructorExp {
	line := lexer.Line()
	lexer.NextTokenOfKind(LEX_SEP_LCURLY)     // {
	keyExps, valExps := parseFieldList(lexer) // [fieldlist]
	lexer.NextTokenOfKind(LEX_SEP_RCURLY)     // }
	lastLine := lexer.Line()
	return &TableConstructorExp{line, lastLine, keyExps, valExps}
}

// fieldlist ::= field {fieldsep field} [fieldsep]
func parseFieldList(lexer *Lexer) (ks, vs []Exp) {
	if lexer.LookAhead() != LEX_SEP_RCURLY {
		k, v := parseField(lexer)
		ks = append(ks, k)
		vs = append(vs, v)

		for isFieldSep(lexer.LookAhead()) {
			lexer.NextToken()
			if lexer.LookAhead() != LEX_SEP_RCURLY {
				k, v := parseField(lexer)
				ks = append(ks, k)
				vs = append(vs, v)
			} else {
				break
			}
		}
	}
	return
}

// fieldsep ::= ‘,’ | ‘;’
func isFieldSep(tokenKind int) bool {
	return tokenKind == LEX_SEP_COMMA || tokenKind == LEX_SEP_SEMI
}

// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
func parseField(lexer *Lexer) (k, v Exp) {
	if lexer.LookAhead() == LEX_SEP_LBRACK {
		lexer.NextToken()                     // [
		k = parseExp(lexer)                   // exp
		lexer.NextTokenOfKind(LEX_SEP_RBRACK) // ]
		lexer.NextTokenOfKind(LEX_OP_ASSIGN)  // =
		v = parseExp(lexer)                   // exp
		return
	}
	exp := parseExp(lexer)
	if nameExp, ok := exp.(*NameExp); ok {
		if lexer.LookAhead() == LEX_OP_ASSIGN {
			// Name ‘=’ exp => ‘[’ LiteralString ‘]’ = exp
			lexer.NextToken()
			k = &StringExp{nameExp.Line, nameExp.Name}
			v = parseExp(lexer)
			return
		}
	}
	return nil, exp
}

// '(' exp ')'
func parseParensExp(lexer *Lexer) Exp {
	lexer.NextTokenOfKind(LEX_SEP_LPAREN) // (
	exp := parseExp(lexer)                // exp
	lexer.NextTokenOfKind(LEX_SEP_RPAREN) // )

	switch exp.(type) {
	case *VarargExp, *FuncCallExp, *NameExp, *TableAccessExp:
		return &ParensExp{exp}
	}

	// no need to keep parens
	return exp
}

// prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name | prefixexp [‘:’ Name] args
func finishPrefixExp(lexer *Lexer, exp Exp) Exp {
	for {
		switch lexer.LookAhead() {
		case LEX_SEP_LBRACK: // prefixexp ‘[’ exp ‘]’
			lexer.NextToken()                     // ‘[’
			keyExp := parseExp(lexer)             // exp
			lexer.NextTokenOfKind(LEX_SEP_RBRACK) // ‘]’
			exp = &TableAccessExp{lexer.Line(), exp, keyExp}
		case LEX_SEP_DOT: // prefixexp ‘.’ Name
			lexer.NextToken()                    // ‘.’
			line, name := lexer.NextIdentifier() // Name
			keyExp := &StringExp{line, name}
			exp = &TableAccessExp{line, exp, keyExp}
		case LEX_SEP_COLON, // prefixexp ‘:’ Name args
			LEX_SEP_LPAREN, LEX_SEP_LCURLY, LEX_STRING: // prefixexp args
			exp = finishFuncCallExp(lexer, exp)
		default:
			return exp
		}
	}
	return exp
}

// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
func finishFuncCallExp(lexer *Lexer, prefixExp Exp) *FuncCallExp {
	nameExp := parseNameExp(lexer)
	line := lexer.Line() // todo
	args := parseArgs(lexer)
	lastLine := lexer.Line()
	return &FuncCallExp{line, lastLine, prefixExp, nameExp, args}
}

// prefixexp ‘:’ Name args
func parseNameExp(lexer *Lexer) *StringExp {
	if lexer.LookAhead() == LEX_SEP_COLON {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		return &StringExp{line, name}
	}
	return nil
}

// args ::=  ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func parseArgs(lexer *Lexer) (args []Exp) {
	switch lexer.LookAhead() {
	case LEX_SEP_LPAREN: // ‘(’ [explist] ‘)’
		lexer.NextToken() // LEX_SEP_LPAREN
		if lexer.LookAhead() != LEX_SEP_RPAREN {
			args = parseExpList(lexer)
		}
		lexer.NextTokenOfKind(LEX_SEP_RPAREN)
	case LEX_SEP_LCURLY: // ‘{’ [fieldlist] ‘}’
		args = []Exp{parseTableConstructorExp(lexer)}
	default: // LiteralString
		line, str := lexer.NextTokenOfKind(LEX_STRING)
		args = []Exp{&StringExp{line, str}}
	}
	return
}
