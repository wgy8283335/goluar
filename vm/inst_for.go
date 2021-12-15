package vm

import (
	. "goluar/api"
	. "goluar/common"
)

/*
	for i = 1,5,20 do f() end
	1	LOADK		0	-1
	2	LOADK		1	-2
	3	LOADK		2	-3
	4	FORPREP 	0	 2
	5	GETTABUP	4	 0	-4	//get upvalue at 0 index in upvalues
	6	CALL		4	 1	 1
	7	FORLOOP		0	-3
	8	RETURN		0	 1
	constants
	1	1
	2	5
	3	20
	4	"f"
	locals
	0	for index
	1	for limit
	2	for step
	4	i
	upvalues
	0	_ENV	1	0
*/

/*
	@description
		FORPREP
		R(A)-=R(A+2); pc+=sBx
			----------------		----------------
		A+3		i:						i:
			----------------		----------------
		A+2		(step):2				(step):2
			----------------	=>	----------------
		A+1		(limit):100				(limit):100
			----------------		----------------
		A		(index):1				(index):-1
			----------------		----------------
*/
func forPrep(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1 // register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	// Transfer type from string to number
	if vm.Type(a) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a))
		vm.Replace(a)
	}
	if vm.Type(a+1) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a + 1))
		vm.Replace(a + 1)
	}
	if vm.Type(a+2) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a + 2))
		vm.Replace(a + 2)
	}
	//R(A)-=R(A+2), prepare the initial index.
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(LUA_OPSUB)
	vm.Replace(a)
	//pc+=sBx, This jump sBx steps to FORLOOP.
	vm.AddPC(sBx)
}

/*
	@description
		FORLOOP
		R(A)+=R(A+2);
		if R(A) <?= R(A+1) then {
			pc+=sBx; R(A+3)=R(A)
		}
			----------------		----------------
		A+3		i:						i:
			----------------		----------------
		A+2		(step):2				(step):2
			----------------	=>	----------------
		A+1		(limit):100				(limit):100
			----------------		----------------
		A		(index):-1				(index):1
			----------------		----------------
*/
func forLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1
	// R(A)+=R(A+2); then index value is 1.
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(LUA_OPADD)
	vm.Replace(a)
	// Calculate a regster and compare the limit. Step could be positive or negative.
	isPositiveStep := vm.ToNumber(a+2) >= 0
	if isPositiveStep && vm.Compare(a, a+1, LUA_OPLE) ||
		!isPositiveStep && vm.Compare(a+1, a, LUA_OPLE) {
		// pc+=sBx, This jump sBx steps back to FORREP.
		// R(A+3)=R(A), Assign new value to i.
		vm.AddPC(sBx)
		vm.Copy(a, a+3)
	}
}

/*
	@description
	TFORLOOP
	if R(A+1) ~= nil then {
		R(A)=R(A+1); pc += sBx
	}
	TFORCALL
	R(A+3),R(A+2+C) := R(A)R(R(A+1),R(A+2));

	If key (R(a+1)) is not nil, copy R(A+1) to R(A). This means get the next key and put in control variable for next loop.
*/
func tForLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sBx)
	}
}
