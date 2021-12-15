package vm

import (
	. "goluar/api"
	. "goluar/common"
)

/* arith */
func instadd(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPADD) } // +
func instsub(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPSUB) } // -
func instmul(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPMUL) } // *
func instmod(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPMOD) } // %
func instpow(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPPOW) } // ^
func instdiv(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPDIV) } // /
func instunm(i Instruction, vm LuaVM) { _unaryArith(i, vm, LUA_OPUNM) }  // -

// R(A) := RK(B) op RK(C)
func _binaryArith(i Instruction, vm LuaVM, op ArithOp) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b)
	vm.GetRK(c)
	vm.Arith(op)
	vm.Replace(a)
}

// R(A) := op R(B)
func _unaryArith(i Instruction, vm LuaVM, op ArithOp) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.PushValue(b)
	vm.Arith(op)
	vm.Replace(a)
}

/* compare */
func eq(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPEQ) } // ==
func lt(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPLT) } // <
func le(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPLE) } // <=

/*
	@description
		if ((RK(B) op RK(C)) ~= A) then pc++
*/
func _compare(i Instruction, vm LuaVM, op CompareOp) {
	a, b, c := i.ABC()
	vm.GetRK(b) //Get value in register b, then push the value in the stack
	vm.GetRK(c) //Get value in register c, then push the value in the stack
	// Compare value at the -1(c) and -2(b) index of the stack. The result is bool
	// Compare value in the register a with 0. The result is bool.
	// Compare the first result with the second result.
	if vm.Compare(-2, -1, op) != (a != 0) {
		vm.AddPC(1) //Auto increment pc.
	}
	vm.Pop(2) // Clean up top 2 values in the stack.
}

/* logical */

// R(A) := not R(B)
func not(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	//Change value in the b register to boolean. Only nil and false should be false, otherwise ture.
	//Put the value at the top of the stack.
	vm.PushBoolean(!vm.ToBoolean(b))
	vm.Replace(a) // Assign the value at the top of the stack to the register A.
}

/*
	@description
		if not (R(A) == C) then pc++
*/
func test(i Instruction, vm LuaVM) {
	a, _, c := i.ABC()
	a += 1

	if vm.ToBoolean(a) != (c != 0) { //Compare the value in b register with oprand C. Cast them to bool and then compare.
		vm.AddPC(1)
	}
}

/*
	@description
		Test and Set.
		if (R(B) == C) then R(A) := R(B) else pc++
*/
func testSet(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	if vm.ToBoolean(b) == (c != 0) { //Compare the value in b register with oprand C. Cast them to bool and then compare.
		vm.Copy(b, a) // Assign the value in B register to A register.
	} else {
		vm.AddPC(1)
	}
}

/* len & concat */

/*
	@decription
		Get length of value in register B.
		R(A) := length of R(B)
*/
func length(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Len(b) //Get the length of b, and push the value into the stack.
	vm.Replace(a)
}

/*
	@decription
		Concat value in register B and C, then assign the result to register A.
 		R(A) := R(B).. ... ..R(C)
*/
func concat(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1

	n := c - b + 1
	vm.CheckStack(n)
	for i := b; i <= c; i++ {
		vm.PushValue(i)
	}
	vm.Concat(n)
	vm.Replace(a)
}
