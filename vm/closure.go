package vm

import (
	. "goluar/api"
	common "goluar/common"
)

type upvalue struct {
	val *luaValue
}

type closure struct {
	proto  *common.FuncProto // lua closure
	goFunc GoFunction        // go closure
	upvals []*upvalue
}

/*
	@description
		Get Upvalues from function proto, and then assign the Upvalues to a closure upvals.
		Return closure.
*/
func newLuaClosure(proto *common.FuncProto) *closure {
	c := &closure{proto: proto}
	// if nUpvals := len(proto.Upvalues); nUpvals > 0 {
	// 	c.upvals = make([]*upvalue, nUpvals)
	// }
	return c
}

/*
	@description
		Use Gofunction f to initialize go closuer c, and then initialize upvals in c.
*/
func newGoClosure(f GoFunction, nUpvals int) *closure {
	c := &closure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}
