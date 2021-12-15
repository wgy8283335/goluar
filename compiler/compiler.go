/*
	See Copyright Notice at LICENSE file.
*/

package compiler

import . "goluar/common"

/*
	@description
		Compile source code file to function proto type.
	@param
		codes	 string	"source codes"
		fileName string	"The name of source code file."
	@return
		proto	 FuncProto	"Function proto type."
*/
func Compile(codes, fileName string) *FuncProto {
	ast := Parse(codes, fileName)
	proto := GenProto(ast)
	setSource(proto, fileName)
	return proto
}

// Set file name in the function FuncProto and sub function protos.
func setSource(proto *FuncProto, fileName string) {
	proto.Source = fileName
	for _, f := range proto.Protos {
		setSource(f, fileName)
	}
}
