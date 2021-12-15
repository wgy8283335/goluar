/*
	See Copyright Notice at LICENSE file.
*/

package compiler

import . "goluar/common"

func toProto(fi *funcInfo) *FuncProto {
	proto := &FuncProto{
		StartLine:    uint32(fi.line),
		EndLine:      uint32(fi.lastLine),
		UpvalueCount: byte(len(fi.upvalues)),
		NumParams:    byte(fi.numParams),
		MaxStackSize: byte(fi.maxRegs),
		Instructions: fi.insts,
		Constants:    getConstants(fi),
		Protos:       toProtos(fi.subFuncs),
	}

	if fi.line == 0 {
		proto.EndLine = 0
	}
	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2 // todo
	}
	if fi.isVararg {
		proto.IsVararg = 1 // todo
	}

	return proto
}

func toProtos(fis []*funcInfo) []*FuncProto {
	protos := make([]*FuncProto, len(fis))
	for i, fi := range fis {
		protos[i] = toProto(fi)
	}
	return protos
}

func getConstants(fi *funcInfo) []interface{} {
	consts := make([]interface{}, len(fi.constants))
	for k, idx := range fi.constants {
		consts[idx] = k
	}
	return consts
}

func getLocVars(fi *funcInfo) []LocVar {
	locVars := make([]LocVar, len(fi.locVars))
	for i, locVar := range fi.locVars {
		locVars[i] = LocVar{
			VarName: locVar.name,
			StartPC: uint32(locVar.startPC),
			EndPC:   uint32(locVar.endPC),
		}
	}
	return locVars
}

func getUpvalueNames(fi *funcInfo) []string {
	names := make([]string, len(fi.upvalues))
	for name, uv := range fi.upvalues {
		names[uv.index] = name
	}
	return names
}
