/*
	See Copyright Notice at LICENSE file
*/
package compiler

var statEmpty = &EmptyStat{}

/*
	@description
		Describe syntax in EBNF:
		stat ::=  ‘;’
			| break
			| ‘::’ Name ‘::’
			| goto Name
			| do block end
			| while exp do block end
			| repeat block until exp
			| if exp then block {elseif exp then block} [else block] end
			| for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
			| for namelist in explist do block end
			| function funcname funcbody
			| local function Name funcbody
			| local namelist [‘=’ explist]
			| varlist ‘=’ explist
			| functioncall
	@param
		lexer	Lexer	"Lexical analyzer"
	@return
		stat	Stat	"The stat struct defined in ast_stat.go"
*/
func parseStat(lexer *Lexer) Stat {
	switch lexer.LookAhead() {
	case LEX_SEP_SEMI:
		return parseEmptyStat(lexer)
	case LEX_KW_BREAK:
		return parseBreakStat(lexer)
	case LEX_KW_DO:
		return parseDoStat(lexer)
	case LEX_KW_WHILE:
		return parseWhileStat(lexer)
	case LEX_KW_REPEAT:
		return parseRepeatStat(lexer)
	case LEX_KW_IF:
		return parseIfStat(lexer)
	case LEX_KW_FOR:
		return parseForStat(lexer)
	case LEX_KW_FUNCTION:
		return parseFuncDefStat(lexer)
	case LEX_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(lexer)
	default:
		return parseAssignOrFuncCallStat(lexer)
	}
}

// ;
func parseEmptyStat(lexer *Lexer) *EmptyStat {
	lexer.NextTokenOfKind(LEX_SEP_SEMI)
	return statEmpty
}

// break
func parseBreakStat(lexer *Lexer) *BreakStat {
	lexer.NextTokenOfKind(LEX_KW_BREAK)
	return &BreakStat{lexer.Line()}
}

// do block end
func parseDoStat(lexer *Lexer) *DoStat {
	lexer.NextTokenOfKind(LEX_KW_DO)  // do
	block := parseBlock(lexer)        // block
	lexer.NextTokenOfKind(LEX_KW_END) // end
	return &DoStat{block}
}

// while exp do block end
func parseWhileStat(lexer *Lexer) *WhileStat {
	lexer.NextTokenOfKind(LEX_KW_WHILE) // while
	exp := parseExp(lexer)              // exp
	lexer.NextTokenOfKind(LEX_KW_DO)    // do
	block := parseBlock(lexer)          // block
	lexer.NextTokenOfKind(LEX_KW_END)   // end
	return &WhileStat{exp, block}
}

// repeat block until exp
func parseRepeatStat(lexer *Lexer) *RepeatStat {
	lexer.NextTokenOfKind(LEX_KW_REPEAT) // repeat
	block := parseBlock(lexer)           // block
	lexer.NextTokenOfKind(LEX_KW_UNTIL)  // until
	exp := parseExp(lexer)               // exp
	return &RepeatStat{block, exp}
}

// if exp then block {elseif exp then block} [else block] end
func parseIfStat(lexer *Lexer) *IfStat {
	exps := make([]Exp, 0, 4)
	blocks := make([]*Block, 0, 4)

	lexer.NextTokenOfKind(LEX_KW_IF)           // if
	exps = append(exps, parseExp(lexer))       // exp
	lexer.NextTokenOfKind(LEX_KW_THEN)         // then
	blocks = append(blocks, parseBlock(lexer)) // block

	for lexer.LookAhead() == LEX_KW_ELSEIF {
		lexer.NextToken()                          // elseif
		exps = append(exps, parseExp(lexer))       // exp
		lexer.NextTokenOfKind(LEX_KW_THEN)         // then
		blocks = append(blocks, parseBlock(lexer)) // block
	}

	// else block => elseif true then block
	if lexer.LookAhead() == LEX_KW_ELSE {
		lexer.NextToken()                           // else
		exps = append(exps, &TrueExp{lexer.Line()}) //
		blocks = append(blocks, parseBlock(lexer))  // block
	}

	lexer.NextTokenOfKind(LEX_KW_END) // end
	return &IfStat{exps, blocks}
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
// for namelist in explist do block end
func parseForStat(lexer *Lexer) Stat {
	lineOfFor, _ := lexer.NextTokenOfKind(LEX_KW_FOR)
	_, name := lexer.NextIdentifier()
	if lexer.LookAhead() == LEX_OP_ASSIGN {
		return finishForNumStat(lexer, lineOfFor, name)
	} else {
		return finishForInStat(lexer, name)
	}
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
func finishForNumStat(lexer *Lexer, lineOfFor int, varName string) *ForNumStat {
	lexer.NextTokenOfKind(LEX_OP_ASSIGN) // for name =
	initExp := parseExp(lexer)           // exp
	lexer.NextTokenOfKind(LEX_SEP_COMMA) // ,
	limitExp := parseExp(lexer)          // exp

	var stepExp Exp
	if lexer.LookAhead() == LEX_SEP_COMMA {
		lexer.NextToken()         // ,
		stepExp = parseExp(lexer) // exp
	} else {
		stepExp = &IntegerExp{lexer.Line(), 1}
	}

	lineOfDo, _ := lexer.NextTokenOfKind(LEX_KW_DO) // do
	block := parseBlock(lexer)                      // block
	lexer.NextTokenOfKind(LEX_KW_END)               // end

	return &ForNumStat{lineOfFor, lineOfDo,
		varName, initExp, limitExp, stepExp, block}
}

// for namelist in explist do block end
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
func finishForInStat(lexer *Lexer, name0 string) *ForInStat {
	nameList := finishNameList(lexer, name0)        // for namelist
	lexer.NextTokenOfKind(LEX_KW_IN)                // in
	expList := parseExpList(lexer)                  // explist
	lineOfDo, _ := lexer.NextTokenOfKind(LEX_KW_DO) // do
	block := parseBlock(lexer)                      // block
	lexer.NextTokenOfKind(LEX_KW_END)               // end
	return &ForInStat{lineOfDo, nameList, expList, block}
}

// namelist ::= Name {‘,’ Name}
func finishNameList(lexer *Lexer, name0 string) []string {
	names := []string{name0}
	for lexer.LookAhead() == LEX_SEP_COMMA {
		lexer.NextToken()                 // ,
		_, name := lexer.NextIdentifier() // Name
		names = append(names, name)
	}
	return names
}

// local function Name funcbody
// local namelist [‘=’ explist]
func parseLocalAssignOrFuncDefStat(lexer *Lexer) Stat {
	lexer.NextTokenOfKind(LEX_KW_LOCAL)
	if lexer.LookAhead() == LEX_KW_FUNCTION {
		return _finishLocalFuncDefStat(lexer)
	} else {
		return finishLocalVarDeclStat(lexer)
	}
}

// function f() end          =>  f = function() end
// function t.a.b.c.f() end  =>  t.a.b.c.f = function() end
// function t.a.b.c:f() end  =>  t.a.b.c.f = function(self) end
// local function f() end    =>  local f; f = function() end
// The statement `local function f () body end` translates to `local f; f = function () body end`
// local function Name funcbody
func _finishLocalFuncDefStat(lexer *Lexer) *LocalFuncDefStat {
	lexer.NextTokenOfKind(LEX_KW_FUNCTION) // local function
	_, name := lexer.NextIdentifier()      // name
	fdExp := parseFuncDefExp(lexer)        // funcbody
	return &LocalFuncDefStat{name, fdExp}
}

// local namelist [‘=’ explist]
func finishLocalVarDeclStat(lexer *Lexer) *LocalVarDeclStat {
	_, name0 := lexer.NextIdentifier()       // local Name
	nameList := finishNameList(lexer, name0) // { , Name }
	var expList []Exp = nil
	if lexer.LookAhead() == LEX_OP_ASSIGN {
		lexer.NextToken()             // ==
		expList = parseExpList(lexer) // explist
	}
	lastLine := lexer.Line()
	return &LocalVarDeclStat{lastLine, nameList, expList}
}

// varlist ‘=’ explist
// functioncall
func parseAssignOrFuncCallStat(lexer *Lexer) Stat {
	prefixExp := parsePrefixExp(lexer)
	if fc, ok := prefixExp.(*FuncCallExp); ok {
		return fc
	} else {
		return parseAssignStat(lexer, prefixExp)
	}
}

// varlist ‘=’ explist |
func parseAssignStat(lexer *Lexer, var0 Exp) *AssignStat {
	varList := finishVarList(lexer, var0) // varlist
	lexer.NextTokenOfKind(LEX_OP_ASSIGN)  // =
	expList := parseExpList(lexer)        // explist
	lastLine := lexer.Line()
	return &AssignStat{lastLine, varList, expList}
}

// varlist ::= var {‘,’ var}
func finishVarList(lexer *Lexer, var0 Exp) []Exp {
	vars := []Exp{checkVar(lexer, var0)}     // var
	for lexer.LookAhead() == LEX_SEP_COMMA { // {
		lexer.NextToken()                         // ,
		exp := parsePrefixExp(lexer)              // var
		vars = append(vars, checkVar(lexer, exp)) //
	} // }
	return vars
}

// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
func checkVar(lexer *Lexer, exp Exp) Exp {
	switch exp.(type) {
	case *NameExp, *TableAccessExp:
		return exp
	}
	lexer.NextTokenOfKind(-1) // trigger error
	panic("unreachable!")
}

// function funcname funcbody
// funcname ::= Name {‘.’ Name} [‘:’ Name]
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
// function t:f (params) body end		--method definition
// function t.f (self, params) body end --function deifnition
// t.f = (self, params) body end		--assign
func parseFuncDefStat(lexer *Lexer) *AssignStat {
	lexer.NextTokenOfKind(LEX_KW_FUNCTION)  // function
	fnExp, hasColon := parseFuncName(lexer) // funcname
	fdExp := parseFuncDefExp(lexer)         // funcbody
	if hasColon {                           // insert self
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
	}

	return &AssignStat{
		LastLine: fdExp.Line,
		VarList:  []Exp{fnExp},
		ExpList:  []Exp{fdExp},
	}
}

// funcname ::= Name {‘.’ Name} [‘:’ Name]
func parseFuncName(lexer *Lexer) (exp Exp, hasColon bool) {
	line, name := lexer.NextIdentifier()
	exp = &NameExp{line, name}

	for lexer.LookAhead() == LEX_SEP_DOT {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		idx := &StringExp{line, name}
		exp = &TableAccessExp{line, exp, idx}
	}
	if lexer.LookAhead() == LEX_SEP_COLON {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		idx := &StringExp{line, name}
		exp = &TableAccessExp{line, exp, idx}
		hasColon = true
	}

	return
}
