/*
	See Copyright Notice at LICENSE file
*/
package compiler

/*
	@description
		Describe syntax in EBNF: block ::= {stat} [retstat].
	@param
		lexer	Lexer	"Lexical analyzer"
	@return
		block	Block	"Block is defined in ast_block.go"
*/
func parseBlock(lexer *Lexer) *Block {
	return &Block{
		Stats:    parseStats(lexer),
		RetExps:  parseRetExps(lexer),
		LastLine: lexer.Line(),
	}
}

// Parse statement until the end of the block.
func parseStats(lexer *Lexer) []Stat {
	stats := make([]Stat, 0, 8)
	for !isReturnOrBlockEnd(lexer.LookAhead()) {
		stat := parseStat(lexer)
		if _, ok := stat.(*EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

// At end of statement(if, for, repeat, function), keywords: end, else, elseif, until, return.
// At end of chunk, keywords: eof.
func isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case LEX_KW_RETURN, LEX_EOF, LEX_KW_END,
		LEX_KW_ELSE, LEX_KW_ELSEIF, LEX_KW_UNTIL:
		return true
	}
	return false
}

// Describe syntax in EBNF:
// retstat ::= return [explist] [‘;’]
// explist ::= exp {‘,’ exp}
func parseRetExps(lexer *Lexer) []Exp {
	if lexer.LookAhead() != LEX_KW_RETURN {
		return nil
	}

	lexer.NextToken()
	switch lexer.LookAhead() {
	case LEX_EOF, LEX_KW_END,
		LEX_KW_ELSE, LEX_KW_ELSEIF, LEX_KW_UNTIL:
		return []Exp{}
	case LEX_SEP_SEMI:
		lexer.NextToken()
		return []Exp{}
	default:
		exps := parseExpList(lexer)
		if lexer.LookAhead() == LEX_SEP_SEMI {
			lexer.NextToken()
		}
		return exps
	}
}
