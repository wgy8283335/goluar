package vm

import (
	. "goluar/api"
)

/*
	obj:f(arg)
	Use SELF to call the method.
	R(A+1) := R(B); R(A) := R(B)[RK(C)]
*/
func self(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.Copy(b, a+1) // copy value in register B to regsiter A
	vm.GetRK(c)     // get constant from consants by c
	vm.GetTable(b)  // get element from table by R(B)[RK(C)]. b is input argument, RK(C) is at the top of stack
	vm.Replace(a)   // assign the obj.f(arg) to register A.
}

/*
	@description
		Initialize a closure by function proto pointed by bx, push the closure to the top of the stack.
		Pop the closure from the stack and assign the closure to register A.
		R(A) := closure(KPROTO[Bx])
*/
func makeClosure(i Instruction, vm LuaVM) {
	a, bx := i.ABx()
	a += 1
	vm.LoadProto(bx)
	vm.Replace(a)
}

/*
	@description
		Load artuments to the top of the stack, and
		Pop the results and assign to the registers from A in the current stack.
		R(A), R(A+1), ..., R(A+B-2) = vararg
*/
func vararg(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b != 1 { // b==0 or b>1
		vm.LoadVararg(b - 1)
		_popResults(a, b, vm)
	}
}

/*
	@description
		R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2));
*/
func tForCall(i Instruction, vm LuaVM) {
	a, _, c := i.ABC()
	a += 1
	//Push function and arguments to the top of the current stack.
	_pushFuncAndArgs(a, 3, vm)
	//Call function. c stands the number of result values. 2 means two arguments. the value of c is 2.
	vm.Call(2, c)
	//Pop the c results and assign to the registers in the current stack from a+3. the value of c is 2.
	//in _popResults(),the actually number of results will be c+1-1,
	_popResults(a+3, c+1, vm)
}

/*
	Reuse the stack of caller function
	return R(A)(R(A+1), ... ,R(A+B-1))
*/
func tailCall(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	// todo: optimize tail call!
	c := 0
	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResults(a, c, vm)
}

/*
	@decription
		Register A store the index of the fucntion.
		Register B store the number of parameters.
		Register C store the number of return values.
		R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
*/
func call(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	//Push function and arguments to the top of the current stack.
	nArgs := _pushFuncAndArgs(a, b, vm)
	//Call function.
	vm.Call(nArgs, c-1)
	//Pop the results and assign to the registers in the current stack.
	_popResults(a, c, vm)
}

/*
	If arguments number b larger than 1.
	Make sure the slots number is big enough for b, otherwise create new slots.
	Get from register i and push to the top of the stack.
	If arguments number b = 0, means the current function receive all of arguments.
	Now some part of arguments stay in the stack, just put the remain arguments and
	function to the top of the stack.
*/
func _pushFuncAndArgs(a, b int, vm LuaVM) (nArgs int) {
	if b >= 1 {
		vm.CheckStack(b)
		for i := a; i < a+b; i++ {
			vm.PushValue(i)
		}
		return b - 1
	} else {
		_fixStack(a, vm)
		return vm.GetTop() - vm.RegisterCount() - 1
	}
}

// Put the remain arguments and function to the top of the stack.
// Rotate a part of the values in the stack.
func _fixStack(a int, vm LuaVM) {
	x := int(vm.ToInteger(-1))
	vm.Pop(1)

	vm.CheckStack(x - a)
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}
	vm.Rotate(vm.RegisterCount()+1, x-a)
}

// C is the number of return values.
// If c = 1, results number is c-1.
// If c > 1, put the result value from the top of stack to the register i.
// if c = 0, means all of result values should return. But we don't pop them from the stack.
// Instead, we put 'a' at the top of the stack, 'a' stands for the registers which should store the return values.
func _popResults(a, c int, vm LuaVM) {
	if c == 1 {
		// no results
	} else if c > 1 {
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else {
		// leave results on stack
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
	}
}

/*
	@description
		Return the results from the register to the top of the stack.
		If b == 1 means no return values;
		If b > 1 get value from register i, and put the result at the top of the stack.
		If b == 0 means some result values are at the top of the stack.
		Rotate the stack to put the rest of the result to the stack.
		return R(A), ... ,R(A+B-2)
*/
func _return(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b == 1 {
		// no return values
	} else if b > 1 {
		// b-1 return values
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		_fixStack(a, vm)
	}
}
