/*
	See Copyright Notice at LICENSE file.
*/
package compiler

/*
	The definition of chunk and block in EBNF.

	chunk ::= block
	block ::= {stat} [retstat]
	retstat ::= return [explist] [‘;’]
	explist ::= exp {‘,’ exp}
*/
type Block struct {
	LastLine int    // The last line number in chunk
	Stats    []Stat // Statements
	RetExps  []Exp  // Return statement
}
