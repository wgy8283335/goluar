package vm

import (
	common "goluar/common"
	"goluar/compiler"
)

// [-0, +1, –]
/*
	Load binary chunk, and initialize _ENV. Put _ENV as upvalues in the current function upvalues.
*/
func (self *luaState) Load(chunk []byte, chunkName, mode string) int {
	var proto *common.FuncProto
	if common.IsBinaryChunk(chunk) {
		proto = common.LoadBinaryChunk(chunk)
	} else {
		proto = compiler.Compile(string(chunk), chunkName)
	}

	c := newLuaClosure(proto)
	self.stack.push(c)
	// if len(proto.UpvalueCount) > 0 {
	// 	env := self.registry.get(LUA_RIDX_GLOBALS)
	// 	c.upvals[0] = &upvalue{&env}
	// }
	return common.LUA_OK
}

// [-(nargs+1), +nresults, e]
func (self *luaState) Call(nArgs, nResults int) {
	val := self.stack.get(-(nArgs + 1))

	c, ok := val.(*closure)
	if !ok {
		if mf := getMetafield(val, "__call", self); mf != nil {
			if c, ok = mf.(*closure); ok {
				self.stack.push(val)
				self.Insert(-(nArgs + 2))
				nArgs += 1
			}
		}
	}

	if ok {
		if c.proto != nil {
			self.callLuaClosure(nArgs, nResults, c)
		} else {
			self.callGoClosure(nArgs, nResults, c)
		}
	} else {
		panic("not function!")
	}
}

/*
	@description
		Create a temp stack for calling closure, initialize the temp stack.
		Push the temp stack to the state, the current stack will be replaced by the temp stack.
		Run the closure on the current stack.
		Pop the temp stack from the state, the current stack will be restored.
		Get the result from the temp stack, and push the result to the current stack.
*/
func (self *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	// create new lua stack
	newStack := newLuaStack(nArgs+common.LUA_MINSTACK, self)
	newStack.closure = c

	// pass args, pop func
	// args is a array of luaValue
	if nArgs > 0 {
		args := self.stack.popN(nArgs)
		newStack.pushN(args, nArgs)
	}
	self.stack.pop() //remove go closure from the caller stack

	// run closure
	self.pushLuaStack(newStack)
	r := c.goFunc(self)
	self.popLuaStack()

	// return results
	if nResults != 0 {
		results := newStack.popN(r)
		self.stack.check(len(results))
		self.stack.pushN(results, nResults)
	}
}

/*
	@description
		Create a temp stack for calling closure, initialize the temp stack.
		Push the temp stack to the state, the current stack will be replaced by the temp stack.
		Run the closure on the current stack.
		Pop the temp stack from the state, the current stack will be restored.
		Get the result from the temp stack, and push the result to the current stack.
*/
func (self *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	// create new lua stack
	newStack := newLuaStack(nRegs+common.LUA_MINSTACK, self)
	newStack.closure = c

	// pass args, pop func
	funcAndArgs := self.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams) //except go closure from funcAndArgs,only arguments are pushed to the new stack.
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	// run closure
	self.pushLuaStack(newStack)
	self.runLuaClosure()
	self.popLuaStack()

	// return results
	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		self.stack.check(len(results))
		self.stack.pushN(results, nResults)
	}
}

func (self *luaState) runLuaClosure() {
	for {
		inst := Instruction(self.Fetch())
		inst.Execute(self)
		if inst.Opcode() == common.OP_RETURN {
			break
		}
	}
}

// Calls a function in protected mode.
// http://www.lua.org/manual/5.3/manual.html#lua_pcall
func (self *luaState) PCall(nArgs, nResults, msgh int) (status int) {
	caller := self.stack
	status = common.LUA_ERRRUN

	// catch error
	defer func() {
		if err := recover(); err != nil {
			if msgh != 0 {
				panic(err)
			}
			for self.stack != caller {
				self.popLuaStack()
			}
			self.stack.push(err)
		}
	}()

	self.Call(nArgs, nResults)
	status = common.LUA_OK
	return
}
