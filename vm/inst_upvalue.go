package vm

import . "goluar/api"

/*
	@description
		Get upvalue from upvalue array by index pointed by B.
		upvalue array is stored in the bottom of the stack. The stack contains: function stack, register, upvalue.
		R(A) := UpValue[B]
*/
func getUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Copy(LuaUpvalueIndex(b), a)
}

/*
	@description
		Assign value in the register A to upvalue pointed by B in the upvalue array.
		UpValue[B] := R(A)
*/
func setUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Copy(a, LuaUpvalueIndex(b))
}
