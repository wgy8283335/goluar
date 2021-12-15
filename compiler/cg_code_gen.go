/*
	See Copyright Notice at LICENSE file.
*/

package compiler

import (
	. "goluar/common"
)

func GenProto(chunk *Block) *FuncProto {
	fd := &FuncDefExp{
		LastLine: chunk.LastLine,
		IsVararg: true,
		Block:    chunk,
	}

	fi := newFuncInfo(nil, fd)
	fi.addLocVar("_ENV", 0)
	cgFuncDefExp(fi, fd, 0)
	return toProto(fi.subFuncs[0])
}
