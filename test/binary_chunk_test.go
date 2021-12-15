package test

import (
	"fmt"
	. "goluar/common"
	"io/ioutil"
	"testing"
)

func TestBinaryChunk(t *testing.T) {
	data, err := ioutil.ReadFile("out/hello.out")
	if err != nil {
		panic(err)
	}

	proto := LoadBinaryChunk(data)
	list(proto)
}

func list(f *FuncProto) {
	printHeader(f)
	printCode(f)
	printDetail(f)
	for _, p := range f.Protos {
		list(p)
	}
}

func printHeader(f *FuncProto) {
	funcType := "main"
	if f.StartLine > 0 {
		funcType = "function"
	}

	varargFlag := ""
	if f.IsVararg > 0 {
		varargFlag = "+"
	}

	fmt.Printf("\n%s <%s:%d,%d> (%d instructions)\n",
		funcType, f.Source, f.StartLine, f.EndLine, len(f.Instructions))

	fmt.Printf("%d%s params, %d slots, ",
		f.NumParams, varargFlag, f.MaxStackSize)

	fmt.Printf("%d constants, %d functions\n",
		len(f.Constants), len(f.Protos))
}

func printCode(f *FuncProto) {
	for pc, c := range f.Instructions {
		line := "-"
		fmt.Printf("\t%d\t[%s]\t0x%08X\n", pc+1, line, c)
	}
}

func printDetail(f *FuncProto) {
	fmt.Printf("constants (%d):\n", len(f.Constants))
	for i, k := range f.Constants {
		fmt.Printf("\t%d\t%s\n", i+1, constantToString(k))
	}
}

func constantToString(k interface{}) string {
	switch k.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%t", k)
	case float64:
		return fmt.Sprintf("%g", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case string:
		return fmt.Sprintf("%q", k)
	default:
		return "?"
	}
}
