package test

import (
	"encoding/json"
	. "goluar/compiler"
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	chunkName := "lua/hello.lua"
	data, err := ioutil.ReadFile(chunkName)
	ast := Parse(string(data), chunkName)
	b, err := json.Marshal(ast)
	if err != nil {
		panic(err)
	}
	println(string(b))
}
