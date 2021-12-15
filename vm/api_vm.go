package vm

func (self *luaState) PC() int {
	return self.stack.pc
}

func (self *luaState) AddPC(n int) {
	self.stack.pc += n
}

func (self *luaState) Fetch() uint32 {
	i := self.stack.closure.proto.Instructions[self.stack.pc]
	self.stack.pc++
	return i
}

func (self *luaState) GetConst(idx int) {
	c := self.stack.closure.proto.Constants[idx]
	self.stack.push(c)
}

func (self *luaState) GetRK(rk int) {
	if rk > 0xFF { // constant
		self.GetConst(rk & 0xFF)
	} else { // register
		self.PushValue(rk + 1)
	}
}

func (self *luaState) RegisterCount() int {
	return int(self.stack.closure.proto.MaxStackSize)
}

func (self *luaState) LoadVararg(n int) {
	if n < 0 {
		n = len(self.stack.varargs)
	}

	self.stack.check(n)
	self.stack.pushN(self.stack.varargs, n)
}

/*
	@description
		Get the function proto from the protos by index. Initialize a closure by the proto.
		Push the closure to the stack. Traverse the upvalues of the sub proto.
		If a upvalue is a local variable of the current proto. We assign sub function closure upvals by current function openuv.
		The openuv contains upvalues which is referenced by the sub fuction, and these upvalues is the current fuction variable in the stack.
		Otherwise, assign sub function closure upvals by current function upvals.
*/
func (self *luaState) LoadProto(idx int) {
	stack := self.stack
	subProto := stack.closure.proto.Protos[idx]
	closure := newLuaClosure(subProto)
	stack.push(closure)
	// for i, uvInfo := range subProto.Upvalues {
	// 	uvIdx := int(uvInfo.Idx)
	// 	if uvInfo.Instack == 1 {
	// 		if stack.openuvs == nil {
	// 			stack.openuvs = map[int]*upvalue{}
	// 		}

	// 		if openuv, found := stack.openuvs[uvIdx]; found {
	// 			closure.upvals[i] = openuv
	// 		} else {
	// 			closure.upvals[i] = &upvalue{&stack.slots[uvIdx]}
	// 			stack.openuvs[uvIdx] = closure.upvals[i]
	// 		}
	// 	} else {
	// 		closure.upvals[i] = stack.closure.upvals[uvIdx]
	// 	}
	// }
}

func (self *luaState) CloseUpvalues(a int) {
	for i, openuv := range self.stack.openuvs {
		if i >= a-1 {
			val := *openuv.val
			openuv.val = &val
			delete(self.stack.openuvs, i)
		}
	}
}
