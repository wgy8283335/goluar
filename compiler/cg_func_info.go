/*
	See Copyright Notice at LICENSE file.
*/
package compiler

import . "goluar/common"

var arithAndBitwiseBinops = map[int]int{
	LEX_OP_ADD: OP_ADD,
	LEX_OP_SUB: OP_SUB,
	LEX_OP_MUL: OP_MUL,
	LEX_OP_MOD: OP_MOD,
	LEX_OP_POW: OP_POW,
	LEX_OP_DIV: OP_DIV,
}

/*
	Upvalue information.
	Upvalue is the variable at the outside of a closure, it is captured by the closure.
*/
type upvalInfo struct {
	locVarSlot int // If the upvalue is in the directly outside function, the locVarSlot record its register index.
	upvalIndex int // If the upvalue is in the indirectly outside function, the locVarSlot record its index in the upvalue table.
	index      int // The sequence of upvalues in the function.
}

/*
	Linked list of local variable information
*/
type locVarInfo struct {
	prev     *locVarInfo // the point of the previous locVarInfo node
	name     string      // the name of the variable
	scopeLv  int         // the socpe level of the variable
	slot     int         // the register index of the variable.
	startPC  int
	endPC    int
	captured bool // whether the local variable is captured by a closure.
}

/*
	Function information.
*/
type funcInfo struct {
	parent    *funcInfo              // the outside scope of the fuction.
	subFuncs  []*funcInfo            // sub functions in this function.
	usedRegs  int                    // the amount of registers which is being used.
	maxRegs   int                    // the maxium amount of registers.
	scopeLv   int                    // scope level.
	locVars   []*locVarInfo          // local variables array
	locNames  map[string]*locVarInfo // The mapping from local variable name to local variable information.
	upvalues  map[string]upvalInfo   // The mapping from upvalue name to upvalue information.
	constants map[interface{}]int    // Constant table store variables in the function body. The types of variable are: nil, bool, digit, string.
	breaks    [][]int                // 'break' is used in for, repeat, while sentences.
	insts     []uint32               // Store byte code.
	lineNums  []uint32               // amount of lines.
	line      int                    // line number at the begin of function.
	lastLine  int                    // line number at the end of expression
	numParams int                    // number of parameters
	isVararg  bool                   // whether the number of arguments is variable.
}

/*
	@desciption
		Create a function inforamtion instance.
*/
func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		locVars:   make([]*locVarInfo, 0, 8),
		locNames:  map[string]*locVarInfo{},
		upvalues:  map[string]upvalInfo{},
		constants: map[interface{}]int{},
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		lineNums:  make([]uint32, 0, 8),
		line:      fd.Line,
		lastLine:  fd.LastLine,
		numParams: len(fd.ParList),
		isVararg:  fd.IsVararg,
	}
}

/*
	@desciption
		Get the index of the constant k in the constants table.
	@param
		k		Any		"constant k"
	@return
		index	int		"index of k in the constants"
*/
func (self *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := self.constants[k]; found {
		return idx
	}

	idx := len(self.constants)
	self.constants[k] = idx
	return idx
}

/*
	@desciption
		Allocate a register, add 1 to the amount of registers being used.
	@return
		usedRegs	int		"the amount of registers being used"
*/
func (self *funcInfo) allocReg() int {
	self.usedRegs++
	if self.usedRegs >= 255 {
		panic("function or expression needs too many registers")
	}
	if self.usedRegs > self.maxRegs {
		self.maxRegs = self.usedRegs
	}
	return self.usedRegs - 1
}

/*
	@desciption
		Free a register, minus 1 to the amount of registers being used.
	@return
		usedRegs	int		"the amount of registers being used"
*/
func (self *funcInfo) freeReg() {
	if self.usedRegs <= 0 {
		panic("usedRegs <= 0 !")
	}
	self.usedRegs--
}

/*
	@desciption
		Allocate n registers, add n to the amount of registers being used.
	@param
		n	int		"the amount of registers needed"
	@return
		usedRegs	int		"the amount of registers being used"
*/
func (self *funcInfo) allocRegs(n int) int {
	if n <= 0 {
		panic("n <= 0 !")
	}
	for i := 0; i < n; i++ {
		self.allocReg()
	}
	return self.usedRegs - n
}

/*
	@desciption
		Free n registers, minus n to the amount of registers being used.
	@param
		n	int		"the amount of registers to be freed"
	@return
		usedRegs	int		"the amount of registers being used"
*/
func (self *funcInfo) freeRegs(n int) {
	if n < 0 {
		panic("n < 0 !")
	}
	for i := 0; i < n; i++ {
		self.freeReg()
	}
}

/*
	@desciption
		Enter a new scope. Add 1 to 'scopeLv' variable.
		If breakable is ture, add one element in the breaks.
	@param
		breakable	bool	"wheter the scope is breakable"
*/
func (self *funcInfo) enterScope(breakable bool) {
	self.scopeLv++
	if breakable {
		self.breaks = append(self.breaks, []int{})
	} else {
		self.breaks = append(self.breaks, nil)
	}
}

/*
	@desciption
		When exit the scope, should delete local variables, recycle registers.
	@param
		endPC	int		""
*/
func (self *funcInfo) exitScope(endPC int) {
	pendingBreakJmps := self.breaks[len(self.breaks)-1]
	self.breaks = self.breaks[:len(self.breaks)-1]
	a := self.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := self.pc() - pc
		i := (sBx+MAXARG_sBx)<<14 | a<<6 | OP_JMP
		self.insts[pc] = uint32(i)
	}
	self.scopeLv--
	for _, locVar := range self.locNames {
		if locVar.scopeLv > self.scopeLv { // out of scope
			locVar.endPC = endPC
			self.removeLocVar(locVar)
		}
	}
}

/*
	@desciption
		Remove local variables. Recycle regsters and delete local variables.
	@param
		locVar	*locVarInfo		"the point of local variable"
*/
func (self *funcInfo) removeLocVar(locVar *locVarInfo) {
	self.freeReg()
	if locVar.prev == nil {
		delete(self.locNames, locVar.name)
	} else if locVar.prev.scopeLv == locVar.scopeLv {
		self.removeLocVar(locVar.prev)
	} else {
		self.locNames[locVar.name] = locVar.prev
	}
}

/*
	@desciption
		Add one local variable in the current scope, and return the register index.
	@param
		name 	string	"the variable name"
		startPC	int		""
	@return
		newVar.slot		int 	"the register index"
*/
func (self *funcInfo) addLocVar(name string, startPC int) int {
	newVar := &locVarInfo{
		name:    name,
		prev:    self.locNames[name],
		scopeLv: self.scopeLv,
		slot:    self.allocReg(),
		startPC: startPC,
		endPC:   0,
	}

	self.locVars = append(self.locVars, newVar)
	self.locNames[name] = newVar

	return newVar.slot
}

/*
	@desciption
		Check whether the local varaible name has been bound to a register.
		Return the regiser index.
	@param
		name 	string		"the local vaiable name"
	@return
		locVar.slot		int		"the register index"
*/
func (self *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := self.locNames[name]; found {
		return locVar.slot
	}
	return -1
}

/*
	@desciption
		Add the jmp command which corresponds to 'break' in the nearest loop block.
	@param
		pc	int		""
*/
func (self *funcInfo) addBreakJmp(pc int) {
	for i := self.scopeLv; i >= 0; i-- {
		if self.breaks[i] != nil { // breakable
			self.breaks[i] = append(self.breaks[i], pc)
			return
		}
	}

	panic("<break> at line ? not inside a loop!")
}

/*
	@desciption
		If the name has been bound in upvalues, return the inedx of upvalue.
		Otherwise, find the upvalue in the parent locNames or recursive parent upvalues.
	@param
		name	string		"upvalue name"
	@return

*/
func (self *funcInfo) indexOfUpval(name string) int {
	if upval, ok := self.upvalues[name]; ok {
		return upval.index
	}
	if self.parent != nil {
		if locVar, found := self.parent.locNames[name]; found {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{locVar.slot, -1, idx}
			locVar.captured = true
			return idx
		}
		if uvIdx := self.parent.indexOfUpval(name); uvIdx >= 0 {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{-1, uvIdx, idx}
			return idx
		}
	}
	return -1
}

// Get 'a' from jmp command, and then generate byte code fo JMP.
// 'a' is the register index of the first local variable need to be handle.
func (self *funcInfo) closeOpenUpvals(line int) {
	a := self.getJmpArgA()
	if a > 0 {
		self.emitJmp(line, a, 0)
	}
}

func (self *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := self.maxRegs
	for _, locVar := range self.locNames {
		if locVar.scopeLv == self.scopeLv {
			for v := locVar; v != nil && v.scopeLv == self.scopeLv; v = v.prev {
				if v.captured {
					hasCapturedLocVars = true
				}
				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}
		}
	}
	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

/*
	@description
		Return the index of last byte code command.
	@return
		result	int	"the index of last command"
*/

func (self *funcInfo) pc() int {
	return len(self.insts) - 1
}

/*
	@description
		Fix xBx in the bytecode command pointed by pc.
*/
func (self *funcInfo) fixSbx(pc, sBx int) {
	i := self.insts[pc]
	i = i << 18 >> 18                  // clear sBx
	i = i | uint32(sBx+MAXARG_sBx)<<14 // reset sBx
	self.insts[pc] = i
}

// todo: rename?
func (self *funcInfo) fixEndPC(name string, delta int) {
	for i := len(self.locVars) - 1; i >= 0; i-- {
		locVar := self.locVars[i]
		if locVar.name == name {
			locVar.endPC += delta
			return
		}
	}
}

/*
	@description
		The command format is below:
		31       22       13       5    0
		^--------^--------^-------^-----
		|b=9bits |c=9bits |a=8bits|op=6|
		op = OP_ADD  R(A) := RK(B) + RK(C)
		K means constant, R means register.
		Record the command in insts, record the line number in lineNums.
*/
func (self *funcInfo) emitABC(line, opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	self.insts = append(self.insts, uint32(i))
	self.lineNums = append(self.lineNums, uint32(line))
}

/*
	@description
		The command format is below:
		31       22       13       5    0
		^-----------------^-------^-----
		|    bx=18bits    |a=8bits|op=6|
		op = OP_LOADK  R[A] := K[Bx]
		Record the command in insts, record the line number in lineNums.
*/
func (self *funcInfo) emitABx(line, opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	self.insts = append(self.insts, uint32(i))
	self.lineNums = append(self.lineNums, uint32(line))
}

/*
	@description
		The command format is below:
		31       22       13       5    0
		^-----------------^-------^-----
		|   sbx=18bits    |a=8bits|op=6|
		op = OP_LOADI  R[A] := sBx
		sBx is a sign integer.
		Record the command in insts, record the line number in lineNums.
*/
func (self *funcInfo) emitAsBx(line, opcode, a, b int) {
	i := (b+MAXARG_sBx)<<14 | a<<6 | opcode
	self.insts = append(self.insts, uint32(i))
	self.lineNums = append(self.lineNums, uint32(line))
}

/*
	@description
		The command format is below:
		31       22       13       5    0
		^-------------------------^-----
		|    ax=26bits            |op=6|
		Only 'EXTRAARG' command use this kind of format. 'EXTRAARG' is used as extended command.
		Record the command in insts, record the line number in lineNums.
*/
// func (self *funcInfo) emitAx(line, opcode, ax int) {
// 	i := ax<<6 | opcode
// 	self.insts = append(self.insts, uint32(i))
// 	self.lineNums = append(self.lineNums, uint32(line))
// }

/*
	@description
		Move value from register b to register a.
		r[a] = r[b]
*/
func (self *funcInfo) emitMove(line, a, b int) {
	self.emitABC(line, OP_MOVE, a, b, 0)
}

/*
	@description
		Assign nil to register a, a+1...a+b
		r[a], r[a+1], ..., r[a+b] = nil
*/
func (self *funcInfo) emitLoadNil(line, a, n int) {
	self.emitABC(line, OP_LOADNIL, a, n-1, 0)
}

/*
	@description
		Assign b to register a. If c is ture, auto increment pc.
		r[a] = (bool)b; if (c) pc++
*/
func (self *funcInfo) emitLoadBool(line, a, b, c int) {
	self.emitABC(line, OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (self *funcInfo) emitLoadK(line, a int, k interface{}) {
	idx := self.indexOfConstant(k)
	self.emitABx(line, OP_LOADK, a, idx)
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (self *funcInfo) emitVararg(line, a, n int) {
	self.emitABC(line, OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (self *funcInfo) emitClosure(line, a, bx int) {
	self.emitABx(line, OP_CLOSURE, a, bx)
}

// r[a] = {}
func (self *funcInfo) emitNewTable(line, a, nArr, nRec int) {
	self.emitABC(line, OP_NEWTABLE,
		a, Int2fb(nArr), Int2fb(nRec))
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (self *funcInfo) emitSetList(line, a, b, c int) {
	self.emitABC(line, OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (self *funcInfo) emitGetTable(line, a, b, c int) {
	self.emitABC(line, OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTable(line, a, b, c int) {
	self.emitABC(line, OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (self *funcInfo) emitGetUpval(line, a, b int) {
	self.emitABC(line, OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (self *funcInfo) emitSetUpval(line, a, b int) {
	self.emitABC(line, OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (self *funcInfo) emitGetTabUp(line, a, b, c int) {
	self.emitABC(line, OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTabUp(line, a, b, c int) {
	self.emitABC(line, OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (self *funcInfo) emitCall(line, a, nArgs, nRet int) {
	self.emitABC(line, OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (self *funcInfo) emitTailCall(line, a, nArgs int) {
	self.emitABC(line, OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
func (self *funcInfo) emitReturn(line, a, n int) {
	self.emitABC(line, OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (self *funcInfo) emitSelf(line, a, b, c int) {
	self.emitABC(line, OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (self *funcInfo) emitJmp(line, a, sBx int) int {
	self.emitAsBx(line, OP_JMP, a, sBx)
	return len(self.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (self *funcInfo) emitTest(line, a, c int) {
	self.emitABC(line, OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (self *funcInfo) emitTestSet(line, a, b, c int) {
	self.emitABC(line, OP_TESTSET, a, b, c)
}

func (self *funcInfo) emitForPrep(line, a, sBx int) int {
	self.emitAsBx(line, OP_FORPREP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitForLoop(line, a, sBx int) int {
	self.emitAsBx(line, OP_FORLOOP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitTForCall(line, a, c int) {
	self.emitABC(line, OP_TFORCALL, a, 0, c)
}

func (self *funcInfo) emitTForLoop(line, a, sBx int) {
	self.emitAsBx(line, OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (self *funcInfo) emitUnaryOp(line, op, a, b int) {
	switch op {
	case LEX_OP_NOT:
		self.emitABC(line, OP_NOT, a, b, 0)
	case LEX_OP_LEN:
		self.emitABC(line, OP_LEN, a, b, 0)
	case LEX_OP_UNM:
		self.emitABC(line, OP_UNM, a, b, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (self *funcInfo) emitBinaryOp(line, op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		self.emitABC(line, opcode, a, b, c)
	} else {
		switch op {
		case LEX_OP_EQ:
			self.emitABC(line, OP_EQ, 1, b, c)
		case LEX_OP_NE:
			self.emitABC(line, OP_EQ, 0, b, c)
		case LEX_OP_LT:
			self.emitABC(line, OP_LT, 1, b, c)
		case LEX_OP_GT:
			self.emitABC(line, OP_LT, 1, c, b)
		case LEX_OP_LE:
			self.emitABC(line, OP_LE, 1, b, c)
		case LEX_OP_GE:
			self.emitABC(line, OP_LE, 1, c, b)
		}
		self.emitJmp(line, 0, 1)
		self.emitLoadBool(line, a, 0, 1)
		self.emitLoadBool(line, a, 1, 0)
	}
}
