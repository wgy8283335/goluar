/*
	See Copyright Notice at LICENSE file.
*/

package compiler

/*
	@desciption
		Generate byte codes for a block node by function information.
		1. Generate byte codes for statement in stats of block node.
		2. Generate byte codes for return statement in RetExps of block node.
*/
func cgBlock(fi *funcInfo, node *Block) {
	for _, stat := range node.Stats {
		cgStat(fi, stat)
	}

	if node.RetExps != nil {
		cgRetStat(fi, node.RetExps, node.LastLine)
	}
}

// Generate return byte code command.
// 1. If nExps is 0, generate return command directly.
// 2. If nExps is 1, get name expression or function call expression, and then generate return command.
// 3. If nExps larger than 1, generate expression before genereate return command.
// need think twice??
func cgRetStat(fi *funcInfo, exps []Exp, lastLine int) {
	nExps := len(exps)
	if nExps == 0 {
		fi.emitReturn(lastLine, 0, 0)
		return
	}

	if nExps == 1 {
		if nameExp, ok := exps[0].(*NameExp); ok {
			if r := fi.slotOfLocVar(nameExp.Name); r >= 0 {
				fi.emitReturn(lastLine, r, 1)
				return
			}
		}
		if fcExp, ok := exps[0].(*FuncCallExp); ok {
			r := fi.allocReg()
			cgTailCallExp(fi, fcExp, r)
			fi.freeReg()
			fi.emitReturn(lastLine, r, -1)
			return
		}
	}

	multRet := isVarargOrFuncCall(exps[nExps-1])
	for i, exp := range exps {
		r := fi.allocReg()
		if i == nExps-1 && multRet {
			cgExp(fi, exp, r, -1)
		} else {
			cgExp(fi, exp, r, 1)
		}
	}
	fi.freeRegs(nExps)

	a := fi.usedRegs // correct?
	if multRet {
		fi.emitReturn(lastLine, a, -1)
	} else {
		fi.emitReturn(lastLine, a, nExps)
	}
}
