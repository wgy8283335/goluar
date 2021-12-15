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

/*
	@description
		Get upvalue from register B, the upvalue's type is table, and then get value from the table by keyRK(C).
		Assign the value to R(A).
		R(A) := UpValue[B][RK(C)]
*/
func getTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.GetRK(c)
	vm.GetTable(LuaUpvalueIndex(b))
	vm.Replace(a)
}

/*
	@description
		Get RK(b) as key. Get RK(C) as value. Put the key and value in the table pointed UpValue[A].
		UpValue[A][RK(B)] := RK(C)
*/
func setTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(LuaUpvalueIndex(a))
}
