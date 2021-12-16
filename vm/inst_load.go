package vm

import . "goluar/api"

/*
	@description
		1. push nil into the top of stack.
		2. copy nil at the top of stack into a, a+1, ... ,a+b registers
		3. pop nil from the top of stack.
		R(A), R(A+1), ..., R(A+B) := nil
*/
func loadNil(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1 // register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	vm.PushNil()
	for i := a; i <= a+b; i++ {
		vm.Copy(-1, i)
	}
	vm.Pop(1)
}

/*
	@description
		Assign one bool value to A register. If B is not 0 ,then bool is true. Otherwise, bool is false.
		If C != 0, pc++.
		R(A) := (bool)B; if (C) pc++
*/
func loadBool(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1                 // register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	vm.PushBoolean(b != 0) // push bool value at the top of stack
	vm.Replace(a)          // assign value at the top of stack to A register.
	if c != 0 {
		vm.AddPC(1)
	}
}

/*
	@description
		Load constant value at the Bx index of the constants table to register A.
		R(A) := Kst(Bx)
*/
func loadK(i Instruction, vm LuaVM) {
	a, bx := i.ABx()
	a += 1          // Register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	vm.GetConst(bx) // Get contant from the contants table by bx index, and push the value into the stack.
	vm.Replace(a)   // Pop value from the stack and assign to A register.
}
