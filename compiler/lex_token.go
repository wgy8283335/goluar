/*
	See Copyright Notice at LICENSE file
*/
package compiler

/*
	The kind of lex LEX.
	Mainly three types: separator, operator, keyword.
	Spearators: ';' ',' '.' ':' '(' ')' '[' ']' '{' '}'.
	Operators: '+'，'-'，'*'，'/'，'%'，'^'，'==','~=','>','<','>=','<=','and','not','or','..','#','='.
	Keywords: and,break,do,else,elseif,end,false,for,function,if,in,local,nil,not,or,repeat,return,then,true,until,while.
*/
const (
	LEX_EOF         = iota         // end-of-file
	LEX_IDENTIFIER                 // identifier
	LEX_KW_BREAK                   // break
	LEX_KW_DO                      // do
	LEX_KW_ELSE                    // else
	LEX_KW_ELSEIF                  // elseif
	LEX_KW_END                     // end
	LEX_KW_FALSE                   // false
	LEX_KW_FOR                     // for
	LEX_KW_FUNCTION                // function
	LEX_KW_IF                      // if
	LEX_KW_IN                      // in
	LEX_KW_LOCAL                   // local
	LEX_KW_NIL                     // nil
	LEX_KW_REPEAT                  // repeat
	LEX_KW_RETURN                  // return
	LEX_KW_THEN                    // then
	LEX_KW_TRUE                    // true
	LEX_KW_UNTIL                   // until
	LEX_KW_WHILE                   // while
	LEX_NUMBER                     // number literal
	LEX_SEP_SEMI                   // ;
	LEX_SEP_COMMA                  // ,
	LEX_SEP_DOT                    // .
	LEX_SEP_COLON                  // :
	LEX_SEP_LPAREN                 // (
	LEX_SEP_RPAREN                 // )
	LEX_SEP_LBRACK                 // [
	LEX_SEP_RBRACK                 // ]
	LEX_SEP_LCURLY                 // {
	LEX_SEP_RCURLY                 // }
	LEX_STRING                     // string literal
	LEX_VARARG                     // ...
	LEX_OP_ASSIGN                  // =
	LEX_OP_MINUS                   // - (sub or unm)
	LEX_OP_ADD                     // +
	LEX_OP_MUL                     // *
	LEX_OP_DIV                     // /
	LEX_OP_POW                     // ^
	LEX_OP_MOD                     // %
	LEX_OP_CONCAT                  // ..
	LEX_OP_LT                      // <
	LEX_OP_LE                      // <=
	LEX_OP_GT                      // >
	LEX_OP_GE                      // >=
	LEX_OP_EQ                      // ==
	LEX_OP_NE                      // ~=
	LEX_OP_LEN                     // #
	LEX_OP_AND                     // and
	LEX_OP_OR                      // or
	LEX_OP_NOT                     // not
	LEX_OP_UNM      = LEX_OP_MINUS // unary minus
	LEX_OP_SUB      = LEX_OP_MINUS
)
