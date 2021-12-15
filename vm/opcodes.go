package vm

import (
	"goluar/api"
	. "goluar/common"
)

type opcode struct {
	testFlag byte // operator is a test (next instruction must be a jump)
	setAFlag byte // instruction set register A
	argBMode byte // B arg mode
	argCMode byte // C arg mode
	opMode   byte // op mode
	name     string
	action   func(i Instruction, vm api.LuaVM)
}

var opcodes = []opcode{
	/*     T  A    B       C     mode         name       action */
	opcode{0, 1, OpArgR, OpArgN, IABC /* */, "MOVE    ", move},        // R(A) := R(B) ---- Copy value from register B to register A.
	opcode{0, 1, OpArgK, OpArgN, IABx /* */, "LOADK   ", loadK},       // R(A) := Kst(Bx) ---- Load constant value at the Bx index of the constants table to register A.
	opcode{0, 1, OpArgN, OpArgN, IABx /* */, "LOADKX  ", loadKx},      // R(A) := Kst(extra arg)---- Load constant value at the Ax index of the constants table to register A. Ax is in the EXTRAARG instuction.LOADKX instuction always followed by EXTRAARG instuction. R(A) := Kst(extra arg)
	opcode{0, 0, OpArgU, OpArgU, IAx /*  */, "EXTRAARG", nil},         // extra (larger) argument for previous opcode ----For example, LOADKX is followed by EXTRAARG. LOADKX use Ax in EXTRAARG to load constant value at the Ax index of the constants table.
	opcode{0, 1, OpArgU, OpArgU, IABC /* */, "LOADBOOL", loadBool},    // R(A) := (bool)B; if (C) pc++  ---- Assign one bool value to A register. If B is not 0 ,then bool is true. Otherwise, bool is false.
	opcode{0, 1, OpArgU, OpArgN, IABC /* */, "LOADNIL ", loadNil},     // R(A), R(A+1), ..., R(A+B) := nil ---- 1. push nil into the top of stack.	2. copy nil at the top of stack into a, a+1, ... ,a+b registers	3. pop nil from the top of stack.
	opcode{0, 1, OpArgR, OpArgK, IABC /* */, "GETTABLE", getTable},    // R(A) := R(B)[RK(C)] ---- Get index from c register or constant.And then get the value from table by the index.Assign the value to A regster.
	opcode{0, 0, OpArgK, OpArgK, IABC /* */, "SETTABLE", setTable},    // R(A)[RK(B)] := RK(C) ---- Assign value in RK(C) to key in RK(B) at R(A) table.
	opcode{0, 1, OpArgU, OpArgU, IABC /* */, "NEWTABLE", newTable},    // R(A) := {} (size = B,C) ---- Create Table by initialize array size B and map size C. Assign to A.
	opcode{0, 1, OpArgR, OpArgK, IABC /* */, "SELF    ", self},        // R(A+1) := R(B); R(A) := R(B)[RK(C)] ---- Use SELF to call the method,copy value(obj) in register B to regsiter A+1.Get element from table by R(B)[RK(C)] (obj.f), and assign to register A. b is input argument, RK(C) is constant pointed by C.
	opcode{0, 1, OpArgK, OpArgK, IABC /* */, "ADD     ", instadd},     // R(A) := RK(B) + RK(C) ---- add
	opcode{0, 1, OpArgK, OpArgK, IABC /* */, "SUB     ", instsub},     // R(A) := RK(B) - RK(C) ---- minus
	opcode{0, 1, OpArgK, OpArgK, IABC /* */, "MUL     ", instmul},     // R(A) := RK(B) * RK(C) ---- multiply
	opcode{0, 1, OpArgK, OpArgK, IABC /* */, "MOD     ", instmod},     // R(A) := RK(B) % RK(C) ---- Mod
	opcode{0, 1, OpArgK, OpArgK, IABC /* */, "POW     ", instpow},     // R(A) := RK(B) ^ RK(C) ---- Exponentiation
	opcode{0, 1, OpArgK, OpArgK, IABC /* */, "DIV     ", instdiv},     // R(A) := RK(B) / RK(C) ---- devide
	opcode{0, 1, OpArgR, OpArgN, IABC /* */, "NOT     ", not},         // R(A) := not R(B) ---- not
	opcode{0, 1, OpArgR, OpArgN, IABC /* */, "LEN     ", length},      // R(A) := length of R(B) ---- #, the lentgh of string or table
	opcode{0, 1, OpArgR, OpArgR, IABC /* */, "CONCAT  ", concat},      // R(A) := R(B).. ... ..R(C) ---- concat string in register b to string in resigter c
	opcode{0, 0, OpArgR, OpArgN, IAsBx /**/, "JMP     ", jmp},         // pc+=sBx; if (A) close all upvalues >= R(A - 1)
	opcode{1, 0, OpArgK, OpArgK, IABC /* */, "EQ      ", eq},          // if ((RK(B) == RK(C)) ~= A) then pc++ ---- the result of RK(B) equal RK(C), compuare with A. Then pc auto-increment.
	opcode{1, 0, OpArgK, OpArgK, IABC /* */, "LT      ", lt},          // if ((RK(B) <  RK(C)) ~= A) then pc++ ---- the result of RK(B) less than RK(C), compuare with A. Then pc auto-increment.
	opcode{1, 0, OpArgK, OpArgK, IABC /* */, "LE      ", le},          // if ((RK(B) <= RK(C)) ~= A) then pc++ ---- the result of RK(B) less than or equal RK(C), compuare with A. Then pc auto-increment.
	opcode{1, 0, OpArgN, OpArgU, IABC /* */, "TEST    ", test},        // if not (R(A) <=> C) then pc++ ---- Compare the value in b register with oprand C. Cast them to bool and then compare. Then pc auto-increment.
	opcode{1, 1, OpArgR, OpArgU, IABC /* */, "TESTSET ", testSet},     // if (R(B) <=> C) then R(A) := R(B) else pc++ ---- Compare the value in b register with oprand C. Cast them to bool and then compare.If true,Assign the value in B register to A register.Otherwise, pc auto-increment.
	opcode{0, 1, OpArgU, OpArgU, IABC /* */, "CALL    ", call},        // R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1)) ---- Register A store the index of the fucntion.	Register B store the number of parameters.	Register C store the number of return values. Final return values store in from register A to A+C-2.
	opcode{0, 1, OpArgU, OpArgN, IABx /* */, "CLOSURE ", makeClosure}, // R(A) := makeClosure(KPROTO[Bx]) ---- Initialize a closure by function proto pointed by bx, push the closure to the top of the stack.	Pop the closure from the stack and assign the closure to register A.
	opcode{0, 1, OpArgU, OpArgU, IABC /* */, "TAILCALL", tailCall},    // return R(A)(R(A+1), ... ,R(A+B-1)) ---- Reuse the stack of caller function
	opcode{0, 0, OpArgU, OpArgN, IABC /* */, "RETURN  ", _return},     // return R(A), ... ,R(A+B-2) ---- Return the results from the register to the top of the stack.If b == 1 means no return values;If b > 1 get value from register i, and put the result at the top of the stack.If b == 0 means some result values are at the top of the stack.Rotate the stack to put the rest of the result to the stack.
	opcode{0, 1, OpArgU, OpArgN, IABC /* */, "VARARG  ", vararg},      // R(A), R(A+1), ..., R(A+B-2) = vararg ---- Load arguments to the top of the stack, and pop the results and assign to the registers from A to A+B-2 in the current stack.
	opcode{0, 0, OpArgU, OpArgU, IABC /* */, "SETLIST ", setList},     // R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B ---- Set list. The list is the array in the table.Put values in registers from R(A+i) to array pointed by R(A),1 <= i <= B
	opcode{0, 1, OpArgU, OpArgN, IABC /* */, "GETUPVAL", getUpval},    // R(A) := UpValue[B] ---- Get upvalue from upvalue array by index pointed by B. Upvalue array is stored in the bottom of the stack. The stack contains: function stack, register, upvalue.
	opcode{0, 1, OpArgU, OpArgK, IABC /* */, "GETTABUP", getTabUp},    // R(A) := UpValue[B][RK(C)] ---- Get upvalue from register B, the upvalue's type is table, and then get value from the table by keyRK(C). Assign the value to R(A).
	opcode{0, 0, OpArgK, OpArgK, IABC /* */, "SETTABUP", setTabUp},    // UpValue[A][RK(B)] := RK(C) ---- Get RK(b) as key. Get RK(C) as value. Put the key and value in the table pointed UpValue[A].
	opcode{0, 0, OpArgU, OpArgN, IABC /* */, "SETUPVAL", setUpval},    // UpValue[B] := R(A) ---- Assign value in the register A to upvalue pointed by B in the upvalue array.
	// for i = 1,5,20 do f() end
	// 1	LOADK		0	-1
	// 2	LOADK		1	-2
	// 3	LOADK		2	-3
	// 4	FORPREP 	0	 2	//this should be 3, to jump to FORLOOP
	// 5	GETTABUP	4	 0	-4	//get upvalue at 0 index in upvalues
	// 6	CALL		4	 1	 1
	// 7	FORLOOP		0	-3	//this should be -2,to jump to GETTABUP
	// 8	RETURN		0	 1
	// @description
	// FORPREP: prepare for sentence, R(A)=R(A)-R(A+2), pc=pc+sBx makes the command move to FORLOOP.
	// R(A)-=R(A+2); pc+=sBx
	// 	----------------		----------------
	// A+3		i:						i:
	// 	----------------		----------------
	// A+2		(step):2				(step):2
	// 	----------------	=>	----------------
	// A+1		(limit):100				(limit):100
	// 	----------------		----------------
	// A		(index):1				(index):-1
	// 	----------------		----------------
	// @description
	// 	FORLOOP: loop sentence. prepare index R(A) to 1. If the step is positive then index should less than limit, otherwise should larger than limit.
	//	Move sBX steps to jump back to the loop
	// 	R(A)+=R(A+2);
	// 	if R(A) <?= R(A+1) then {
	// 		pc+=sBx; R(A+3)=R(A)
	// 	}
	// 		----------------		----------------
	// 	A+3		i:						i:
	// 		----------------		----------------
	// 	A+2		(step):2				(step):2
	// 		----------------	=>	----------------
	// 	A+1		(limit):100				(limit):100
	// 		----------------		----------------
	// 	A		(index):-1				(index):1
	// 		----------------		----------------
	opcode{0, 1, OpArgR, OpArgN, IAsBx /**/, "FORLOOP ", forLoop}, // R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) } ---- R(A):index; R(A+1):limit; R(A+2):step; R(A+3):i; sBx: jump steps
	opcode{0, 1, OpArgR, OpArgN, IAsBx /**/, "FORPREP ", forPrep}, // R(A)-=R(A+2); pc+=sBx //Prepare for sentence. ---- R(A):index; R(A+1):limit; R(A+2):step; R(A+3):i; sBx: jump steps
	// for k,v in pairs(t) do print(k,v) end
	// 1	GETTABLEUP 0 0 -1；_ENV "paris"
	// 2	GETTABLEUP 1 0 -2; _ENV "t"
	// 3	CALL       0 2 4
	// 4	JMP        0 4   ; to 9
	// 5	GETTABLEUP 5 0 -3; _ENV "print"
	// 6 	MOVE       6 3
	// 7	MOVE       7 4
	// 8	CALL       5 3 1
	// 9 	TFORCALL   0 2
	// 10	TFORLOOP   2 -6  ; to 5
	// 11	RETURN     0 1
	// TFORLOOP
	// 	Assign R(A+1) to R(A).And increment pc by sBx to go back loop.
	// 		----------------		----------------
	// 			v						v
	// 		----------------		----------------
	// 	A+1		k			   =|		k
	// 		----------------	|	----------------
	// 	A	(control/key):var	|=>	(control/key):var
	// 		----------------		----------------
	// 		(state/table):s			(state/table):s
	// 		----------------		----------------
	// 		(generator/next):f		(generator/next):f
	// 		----------------		----------------
	opcode{0, 1, OpArgR, OpArgN, IAsBx /**/, "TFORLOOP", tForLoop}, // if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }
	// TFORCALL
	// R(A) store the function. R(A+1 store the table.R(A+2) store the key.
	// R(A+3) store the key. R(A+2+C) store the related value.
	// 		----------------		    		----------------
	// 	A+2+C	v					|=>				v
	// 		----------------		|			----------------
	// 	A+3		k			    	|=>				k
	// 		---------------- 	f(s,var)		----------------
	// 	A+2	(control/key):var  	   =|			(control/key):var
	// 		----------------		|			----------------
	// 	A+1	(state/table):s	  	   =|			(state/table):s
	// 		----------------		|			----------------
	// 	A	(generator/next):f	   =|			(generator/next):f
	// 		----------------					----------------
	opcode{0, 0, OpArgN, OpArgU, IABC /* */, "TFORCALL", tForCall}, // R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2));
}
