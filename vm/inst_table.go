package vm

import (
	. "goluar/api"
	. "goluar/common"
)

/* number of list items to accumulate before a SETLIST instruction */
const LFIELDS_PER_FLUSH = 50

/*
	@description
		Transfer value from float point byte to int, b represents array size and c represents map size.
		Push the pointer of the table in stack.Get the pointer from stack and assign to a.
		R(A) := {} (size = B,C)
*/
func newTable(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1 // register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	// Transfer value from float point byte to int, b represents array size and c represents map size.
	//Push the pointer of the table in stack.
	vm.CreateTable(Fb2int(b), Fb2int(c))
	vm.Replace(a) // Get the pointer from stack and assign to a.
}

/*
	@description
		Get index from c register or constant.And then get the value from table by the index.
		Assign the value to A regster.
		R(A) := R(B)[RK(C)]
*/
func getTable(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.GetRK(c)    //Get index from c register or constant, and push to the stack.
	vm.GetTable(b) //Get the table poiter, and the pop the index from the stack. And call getTable().Push the result in the stack.
	vm.Replace(a)  //Pop the value from the stack and assign to a.
}

/*
	@description
		Assign value in RK(C) to key in RK(B) at R(A) table.
		R(A)[RK(B)] := RK(C)
*/
func setTable(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.GetRK(b)    //Get index from b register or constants, and push to the stack.
	vm.GetRK(c)    //Get index from c register or constants, and push to the stack.
	vm.SetTable(a) //a as t,pop b from tack ,pop c from stack. t[b]=c
}

/*
	@description
		Set list. The list is the array in the table.
		Put values in registers from R(A+i) to array pointed by R(A),1 <= i <= B
		For '(C-1)*LFIELDS_PER_FLUSH+i', this stands for the index in array.
		C has only 9 bytes, couldn't satisfy the length of the array.LFIELDS_PER_FLUSH stands for the batch size,
		Use "(C-1)*LFIELDS_PER_FLUSH+i" instead of C to enlarge the value scope.
		A stands for array. B stands for the amount of values to be written in array.
		C stands for the batch number.
		A = A
		C = (C-1)*LFIELDS_PER_FLUSH+i
		B = A+i (1<=i<=B)
		R(A)[(C-1)*LFIELDS_PER_FLUSH+i] := R(A+i), 1 <= i <= B
		We could use EXTRAARG command to store the FPF value in Ax operand.

*/
func setList(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1 // register index should add 1 to transfer to stack index， stack index from 1, register index from 0.
	//If c > 0, the batch number is stored in C as (batch number)+1. if c = 0 , he batch number is stored in next command.
	if c == 0 {
		c = int(vm.Fetch()) //
	}
	c = c - 1
	//Check whether b is zero.
	bIsZero := b == 0
	if bIsZero {
		b = int(vm.ToInteger(-1)) - a - 1
		vm.Pop(1)
	}
	//确保stack中有一个空闲位置
	vm.CheckStack(1)
	//Get batch number according to c, get batch size according to LFIELDS_PER_FLUSH.
	// Then multiply the two values, get the index
	idx := int64(c * LFIELDS_PER_FLUSH)
	for j := 1; j <= b; j++ {
		idx++
		vm.PushValue(a + j) //Get value from register (a+j), push to the stack.
		vm.SetI(a, idx)     //Get table from register a, get index from idx, get value by poping the stack.
	}
	// If b is zero,
	if bIsZero {
		for j := vm.RegisterCount() + 1; j <= vm.GetTop(); j++ {
			idx++
			vm.PushValue(j) //Get value from register (j), push to the stack.
			vm.SetI(a, idx) //Get table from register a, get index from idx, get value by poping the stack.
		}
		//finally, get the values from （RegisterCount() + 1) to vm.GetTop(), and put these value to the array pointed by a.
		// clear stack
		vm.SetTop(vm.RegisterCount()) // restore the top of the stack. clear the values.
	}
}
