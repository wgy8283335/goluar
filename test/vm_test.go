package test

import (
	state "goluar/vm"
	"io/ioutil"
	"os"
	"testing"
)

func TestVM(t *testing.T) {
	data, err := ioutil.ReadFile("out/hello.out")
	if err != nil {
		panic(err)
	}
	ls := state.New()
	ls.Load(data, os.Args[1], "b")
	ls.Call(0, 0)
}
