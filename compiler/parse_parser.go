/*
	See Copyright Notice at LICENSE file
*/
package compiler

/*
	@description
		Parser which parse lua file into ast block.
		1. Create a lexer by specific file.
		2. Parse ast block.
		3. Check whether reach the end of the file.
	@param
		codes		string	"source codes"
		fileName	string 	"the file of source codes"
	@return
		block	Block	"Block is defined in ast_block.go"
*/
func Parse(codes, fileName string) *Block {
	lexer := NewLexer(codes, fileName)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(LEX_EOF)
	return block
}
