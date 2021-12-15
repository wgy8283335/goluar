/*
	Licensed to the Apache Software Foundation (ASF) under one or more
	contributor license agreements.  See the NOTICE file distributed with
	this work for additional information regarding copyright ownership.
	The ASF licenses this file to You under the Apache License, Version 2.0
	(the "License"); you may not use this file except in compliance with
	the License.  You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

 	Unless required by applicable law or agreed to in writing, software
 	distributed under the License is distributed on an "AS IS" BASIS,
 	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 	See the License for the specific language governing permissions and
 	limitations under the License.
*/

package compiler

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
	A lexer of lua.
	This lexer is FSM(Finite-state Machine) kind.
	The line number in source code chunk and the first byte in source code chunk are the inner states of the FSM.
	NextToken() method is implemented by a switch-case sentence. This is the core of FSM.
*/

type Lexer struct {
	codes         string // source codes
	srcFileName   string // source file name
	line          int    // current line number
	nextToken     string // next token
	nextTokenKind int    // next token kind
	nextTokenLine int    // next token line
}

/*
	@description
		Construct the lexer.
	@param
		code	string	"source codes"
		srcFileName	string	"source file name"
	@return
		lexer	Lexer	"Lexical analyzer"
*/
func NewLexer(codes, srcFileName string) *Lexer {
	return &Lexer{codes, srcFileName, 1, "", 0, 0}
}

/*
	@description
		Get the current line number.
	@return
		line	int	"line number"
*/
func (self *Lexer) Line() int {
	return self.line
}

/*
	@description
		Get the kind of next token.
	@return
		kind	int	"token kind defined in lex.token.go"
*/
func (self *Lexer) LookAhead() int {
	// If NextToken() method has been called before LookAhead(),
	// this block will return nextTokenKind directly.
	if self.nextTokenLine > 0 {
		return self.nextTokenKind
	}
	currentLine := self.line
	line, kind, token := self.NextToken()
	self.line = currentLine
	self.nextTokenLine = line
	self.nextTokenKind = kind
	self.nextToken = token
	return kind
}

/*
	@description
		Get next LEX_IDENTIFIER token.
	@return
		line	int		"location of the token"
		token	string	"lex token"
*/
func (self *Lexer) NextIdentifier() (line int, token string) {
	return self.NextTokenOfKind(LEX_IDENTIFIER)
}

/*
	@description
		Get the specified kind of the token, otherwise print error.
	@param
		kink	int		"token kind"
	@return
		line	int		"location of the token"
		token	string	"lex token"

*/
func (self *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, _kind, token := self.NextToken()
	if kind != _kind {
		self.error("syntax error near '%s'", token)
	}
	return line, token
}

/*
	@description
		Get the next token from source code chunk.
		1. Skip white spaces.
		2. Match separator and operator.
		3. Match number.
		4. Match keyword.
		5. Match identifier.
		Return line number, lex token kind and token.
	@return
		kink	int		"token kind"
		line	int		"location of the token"
		token	string	"lex token"
*/
func (self *Lexer) NextToken() (line, kind int, token string) {
	// If LookAhead() method has been calles, this block would be executed
	// as nextTokenLine has been assigned in LookAhead().
	if self.nextTokenLine > 0 {
		line = self.nextTokenLine
		kind = self.nextTokenKind
		token = self.nextToken
		self.line = self.nextTokenLine
		self.nextTokenLine = 0
		return
	}

	self.skipWhiteSpaces()
	if len(self.codes) == 0 {
		return self.line, LEX_EOF, "EOF"
	}

	switch self.codes[0] {
	case ';':
		self.next(1)
		return self.line, LEX_SEP_SEMI, ";"
	case ',':
		self.next(1)
		return self.line, LEX_SEP_COMMA, ","
	case '(':
		self.next(1)
		return self.line, LEX_SEP_LPAREN, "("
	case ')':
		self.next(1)
		return self.line, LEX_SEP_RPAREN, ")"
	case ']':
		self.next(1)
		return self.line, LEX_SEP_RBRACK, "]"
	case '{':
		self.next(1)
		return self.line, LEX_SEP_LCURLY, "{"
	case '}':
		self.next(1)
		return self.line, LEX_SEP_RCURLY, "}"
	case '+':
		self.next(1)
		return self.line, LEX_OP_ADD, "+"
	case '-':
		self.next(1)
		return self.line, LEX_OP_MINUS, "-"
	case '*':
		self.next(1)
		return self.line, LEX_OP_MUL, "*"
	case '^':
		self.next(1)
		return self.line, LEX_OP_POW, "^"
	case '%':
		self.next(1)
		return self.line, LEX_OP_MOD, "%"
	case '#':
		self.next(1)
		return self.line, LEX_OP_LEN, "#"
	case ':':
		self.next(1)
		return self.line, LEX_SEP_COLON, ":"
	case '/':
		self.next(1)
		return self.line, LEX_OP_DIV, "/"
	case '~':
		if self.test("~=") {
			self.next(2)
			return self.line, LEX_OP_NE, "~="
		}
	case '=':
		if self.test("==") {
			self.next(2)
			return self.line, LEX_OP_EQ, "=="
		} else {
			self.next(1)
			return self.line, LEX_OP_ASSIGN, "="
		}
	case '<':
		if self.test("<=") {
			self.next(2)
			return self.line, LEX_OP_LE, "<="
		} else {
			self.next(1)
			return self.line, LEX_OP_LT, "<"
		}
	case '>':
		if self.test(">=") {
			self.next(2)
			return self.line, LEX_OP_GE, ">="
		} else {
			self.next(1)
			return self.line, LEX_OP_GT, ">"
		}
	case '.':
		if self.test("...") {
			self.next(3)
			return self.line, LEX_VARARG, "..."
		} else if self.test("..") {
			self.next(2)
			return self.line, LEX_OP_CONCAT, ".."
		} else if len(self.codes) == 1 || !isDigit(self.codes[1]) {
			self.next(1)
			return self.line, LEX_SEP_DOT, "."
		}
	case '[':
		if self.test("[[") || self.test("[=") {
			return self.line, LEX_STRING, self.scanLongString()
		} else {
			self.next(1)
			return self.line, LEX_SEP_LBRACK, "["
		}
	case '\'', '"':
		return self.line, LEX_STRING, self.scanShortString()
	}

	c := self.codes[0]
	if c == '.' || isDigit(c) {
		token := self.scanNumber()
		return self.line, LEX_NUMBER, token
	}
	if c == '_' || isLetter(c) {
		token := self.scanIdentifier()
		if kind, found := keywords[token]; found {
			return self.line, kind, token // keyword
		} else {
			return self.line, LEX_IDENTIFIER, token
		}
	}

	self.error("unexpected symbol near %q", c)
	return
}

// Jump to the next n position of the source code chunk.
func (self *Lexer) next(n int) {
	self.codes = self.codes[n:]
}

// whether the source code chunk begins with prefix s.
func (self *Lexer) test(s string) bool {
	return strings.HasPrefix(self.codes, s)
}

//Print out error information.
func (self *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	err = fmt.Sprintf("%s:%d: %s", self.srcFileName, self.line, err)
	panic(err)
}

// Skip commnet and whitespace.
// ??考虑换一下方法名和路基判断顺序，for是否可以变成while

func (self *Lexer) skipWhiteSpaces() {
	for len(self.codes) > 0 {
		if self.test("--") {
			self.skipComment()
		} else if self.test("\r\n") || self.test("\n\r") {
			self.next(2)
			self.line += 1
		} else if isNewLine(self.codes[0]) {
			self.next(1)
			self.line += 1
		} else if isWhiteSpace(self.codes[0]) {
			self.next(1)
		} else {
			break
		}
	}
}

// Skip comments, includes: long comment and short comment.
// For long comment, skip '--', then remove '[[', ']]', and long string.
// For shrot comment, skip '--' and string until '\n'.
func (self *Lexer) skipComment() {
	self.next(2) // skip --

	// long comment
	if self.test("[") {
		if reOpeningLongBracket.FindString(self.codes) != "" {
			self.scanLongString()
			return
		}
	}

	// short comment
	for len(self.codes) > 0 && !isNewLine(self.codes[0]) {
		self.next(1)
	}
}

// Get a left most identifier in the source code chunk.
func (self *Lexer) scanIdentifier() string {
	return self.scan(reIdentifier)
}

// Get a left most number in the source code chunk.
func (self *Lexer) scanNumber() string {
	return self.scan(reNumber)
}

// Get a left most sub string by matching regular expression in the source code chunk.
func (self *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(self.codes); token != "" {
		self.next(len(token))
		return token
	}
	panic("unreachable!")
}

// Get long String.
// A long string only appears in the long comment. It is contained by '[[' and ']]'.
// 1. Use regular expression to find '[[' in source code chunk.
// 2. Get the string between openingLongBracket and closingLongBracket.
// 	openingLongBracketn :'[[', closingLongBracket:']]'.
// 3. Replace the "\n" in the string. And jump next n lines by setting self.line vlaue.
// 4. return the string.
func (self *Lexer) scanLongString() string {
	openingLongBracket := reOpeningLongBracket.FindString(self.codes)
	if openingLongBracket == "" {
		self.error("invalid long string delimiter near '%s'",
			self.codes[0:2])
	}

	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIdx := strings.Index(self.codes, closingLongBracket)
	if closingLongBracketIdx < 0 {
		self.error("unfinished long string or comment")
	}

	str := self.codes[len(openingLongBracket):closingLongBracketIdx]
	self.next(closingLongBracketIdx + len(closingLongBracket))

	str = reNewLine.ReplaceAllString(str, "\n")
	self.line += strings.Count(str, "\n")
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}

	return str
}

// Get short string.
// The short string could appears in comments or source codes.
// 1. Use regular expression to find the short string.
// 2. Jump next n step in chunk. The 'n' is the value of string length.
// 3. Jump next n lines. The 'n' is the number of lines.
// 4. Handle escape character.
func (self *Lexer) scanShortString() string {
	if str := reShortStr.FindString(self.codes); str != "" {
		self.next(len(str))
		str = str[1 : len(str)-1]
		if strings.Index(str, `\`) >= 0 {
			self.line += len(reNewLine.FindAllString(str, -1))
			str = self.escape(str)
		}
		return str
	}
	self.error("unfinished string")
	return ""
}

// Handle the escape character in the string.
// 1. Check length and first charactor in the string.
// 2. For \a,\b,\f,\n,\r,\t,\v,\',\",\\ , keep the same value.
// 3. For numbers, like: '\123','\x09', store as integer in buffer.
// 4. For unicode, like: '\u0021', store as rune(int32) in buffer.
func (self *Lexer) escape(str string) string {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			self.error("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				self.error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				self.error("UTF-8 value too large near '%s'", found)
			}
		case 'z':
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { //待考虑是否需要？？
				str = str[1:]
			}
			continue
		}
		self.error("invalid escape sequence near '\\%c'", str[1])
	}

	return buf.String()
}

// Whether the input parameter is white space.
func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

// Whether the input parameter is a new line charactor.

func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

// Whether the input parameter is a digit.
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// Whether the input parameter is a letter.
func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// Define regular expression.
var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reIdentifier = regexp.MustCompile(`^[_\d\w]+`)
var reNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")
var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
var reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

// The map of keywords.
// Mapping from keyword to constant value.
var keywords = map[string]int{
	"and":      LEX_OP_AND,
	"break":    LEX_KW_BREAK,
	"do":       LEX_KW_DO,
	"else":     LEX_KW_ELSE,
	"elseif":   LEX_KW_ELSEIF,
	"end":      LEX_KW_END,
	"false":    LEX_KW_FALSE,
	"for":      LEX_KW_FOR,
	"function": LEX_KW_FUNCTION,
	"if":       LEX_KW_IF,
	"in":       LEX_KW_IN,
	"local":    LEX_KW_LOCAL,
	"nil":      LEX_KW_NIL,
	"not":      LEX_OP_NOT,
	"or":       LEX_OP_OR,
	"repeat":   LEX_KW_REPEAT,
	"return":   LEX_KW_RETURN,
	"then":     LEX_KW_THEN,
	"true":     LEX_KW_TRUE,
	"until":    LEX_KW_UNTIL,
	"while":    LEX_KW_WHILE,
}
