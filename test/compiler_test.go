package test

import (
	. "goluar/compiler"
	"io/ioutil"
	"testing"
)

func TestCompiler(t *testing.T) {
	chunkName := "lua/hello.lua"
	data, err := ioutil.ReadFile(chunkName)
	if err != nil {
		panic(err)
	}
	proto := Compile(string(data), chunkName)
	println("--------------------TestCompiler------------------")
	list(proto)
}
