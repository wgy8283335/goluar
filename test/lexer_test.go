package test

import (
	"fmt"
	. "goluar/compiler"
	"io/ioutil"
	"testing"
)

func TestLexer(t *testing.T) {
	chunkName := "lua/hello.lua"
	data, err := ioutil.ReadFile(chunkName)
	if err != nil {
		panic(err)
	}
	lexer := NewLexer(string(data), chunkName)
	for {
		line, kind, token := lexer.NextToken()
		fmt.Printf("[%2d] [%-10s] %s\n",
			line, kindToCategory(kind), token)
		if kind == LEX_EOF {
			break
		}
	}
}

func kindToCategory(kind int) string {
	switch {
	case (LEX_SEP_SEMI <= kind && kind <= LEX_SEP_RCURLY):
		return "separator"
	case (LEX_OP_ASSIGN <= kind && kind <= LEX_OP_SUB):
		return "operator"
	case (LEX_KW_BREAK <= kind && kind <= LEX_KW_WHILE):
		return "keyword"
	case kind == LEX_IDENTIFIER:
		return "identifier"
	case kind == LEX_NUMBER:
		return "number"
	case kind == LEX_STRING:
		return "string"
	default:
		return "other"
	}
}
