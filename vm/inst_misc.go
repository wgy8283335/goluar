package vm

import . "goluar/api"

/*
	@description
		Copy value from register b to register a.
		R(A) := R(B)
*/
func move(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1 // register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	b += 1 // register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	vm.Copy(b, a)
}

/*
	@description
		Add pc count by sBx, so that the command jump 'sBx' steps.
		pc+=sBx; if (A) close all upvalues >= R(A - 1)
*/
func jmp(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	vm.AddPC(sBx)
	if a != 0 {
		vm.CloseUpvalues(a)
	}
}
