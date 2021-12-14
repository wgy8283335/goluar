package common

import (
	"encoding/binary"
	"math"
)

/*
	Reader of binary chunk which is stored in a byte array.
*/
type reader struct {
	data []byte // store binary chunk
}

/*
	Return 1 byte from data.
*/
func (self *reader) readByte() byte {
	b := self.data[0]
	self.data = self.data[1:]
	return b
}

/*
	Return n byte from data.
*/
func (self *reader) readBytes(n uint) []byte {
	bytes := self.data[:n]
	self.data = self.data[n:]
	return bytes
}

/*
	Return a uint32 value from data.
*/
func (self *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(self.data)
	self.data = self.data[4:]
	return i
}

/*
	Return a uint64 value from data.
*/
func (self *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(self.data)
	self.data = self.data[8:]
	return i
}

/*
	Return a uint64 value from data.
*/
func (self *reader) readLuaInteger() int64 {
	return int64(self.readUint64())
}

/*
	Return a float64 value from data.
*/
func (self *reader) readLuaNumber() float64 {
	return math.Float64frombits(self.readUint64())
}

/*
	Return string from data.
*/
func (self *reader) readString() string {
	size := uint(self.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(self.readUint64()) // size_t
	}
	bytes := self.readBytes(size - 1)
	return string(bytes) // todo
}

/*
	+-----------------------+ ---
	|	signature			|  |
	-------------------------  |
	|	version				|  |
	-------------------------  |
	|	format				|  |
	-------------------------  |
	|	luac data			|  |
	-------------------------  |
	|	cint size			|  |
	-------------------------  | header
	|	csizet size			|  |
	-------------------------  |
	|	instruction size	|  |
	-------------------------  |
	|	lua Integer size	|  |
	-------------------------  |
	|	lua Number size		|  |
	-------------------------  |
	|	luac int			|  |
	-------------------------  |
	|	luac num			|  |
	+-----------------------+ ---
	Check the header of binary chunk, to ensure that the binary chunk is compatible with the lua of current version.
	This is very necessary before parse the binary chunk.
*/
func (self *reader) checkHeader() {
	if string(self.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk!")
	}
	if self.readByte() != LUAC_VERSION {
		panic("version mismatch!")
	}
	if self.readByte() != LUAC_FORMAT {
		panic("format mismatch!")
	}
	if string(self.readBytes(6)) != LUAC_DATA {
		panic("corrupted!")
	}
	if self.readByte() != CINT_SIZE {
		panic("int size mismatch!")
	}
	if self.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch!")
	}
	if self.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch!")
	}
	if self.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch!")
	}
	if self.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch!")
	}
	if self.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch!")
	}
	if self.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch!")
	}
}

/*
	+-----------------------+ ---
	|	Source				|  |
	-------------------------  |
	|	LineDefined			|  |
	-------------------------  |
	|	LastLineDefined		|  |
	-------------------------  |
	|	Numparams			|  |
	-------------------------  |
	|	IsVararg			|  |
	-------------------------  |
	|	MaxStackSize		|  |
	-------------------------  |
	|	Code				|  | function prototype
	-------------------------  |
	|	Constants			|  |
	-------------------------  |
	|	Upvalues			|  |
	-------------------------  |
	|	Protos				|  |
	-------------------------  |
	|	LineInfo			|  |
	-------------------------  |
	|	LocVars				|  |
	-------------------------  |
	|	UpvalueNames		|  |
	+-----------------------+ ---
	Return function proto from binary chunk.
*/
func (self *reader) readProto(parentSource string) *Prototype {
	source := self.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefined:     self.readUint32(),
		LastLineDefined: self.readUint32(),
		NumParams:       self.readByte(),
		IsVararg:        self.readByte(),
		MaxStackSize:    self.readByte(),
		Code:            self.readCode(),
		Constants:       self.readConstants(),
		Upvalues:        self.readUpvalues(),
		Protos:          self.readProtos(source),
		LineInfo:        self.readLineInfo(),
		LocVars:         self.readLocVars(),
		UpvalueNames:    self.readUpvalueNames(),
	}
}

/*
	Return n byte from data.
*/
func (self *reader) readCode() []uint32 {
	code := make([]uint32, self.readUint32())
	for i := range code {
		code[i] = self.readUint32()
	}
	return code
}

/*
	Return n byte from data.
*/
func (self *reader) readConstants() []interface{} {
	constants := make([]interface{}, self.readUint32())
	for i := range constants {
		constants[i] = self.readConstant()
	}
	return constants
}

/*
	Return n byte from data.
*/
func (self *reader) readConstant() interface{} {
	switch self.readByte() {
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return self.readByte() != 0
	case TAG_INTEGER:
		return self.readLuaInteger()
	case TAG_NUMBER:
		return self.readLuaNumber()
	case TAG_SHORT_STR, TAG_LONG_STR:
		return self.readString()
	default:
		panic("corrupted!") // todo
	}
}

/*
	Return n byte from data.
*/
func (self *reader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, self.readUint32())
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: self.readByte(),
			Idx:     self.readByte(),
		}
	}
	return upvalues
}

/*
	Return n byte from data.
*/
func (self *reader) readProtos(parentSource string) []*Prototype {
	protos := make([]*Prototype, self.readUint32())
	for i := range protos {
		protos[i] = self.readProto(parentSource)
	}
	return protos
}

/*
	Return n byte from data.
*/
func (self *reader) readLineInfo() []uint32 {
	lineInfo := make([]uint32, self.readUint32())
	for i := range lineInfo {
		lineInfo[i] = self.readUint32()
	}
	return lineInfo
}

/*
	Return n byte from data.
*/
func (self *reader) readLocVars() []LocVar {
	locVars := make([]LocVar, self.readUint32())
	for i := range locVars {
		locVars[i] = LocVar{
			VarName: self.readString(),
			StartPC: self.readUint32(),
			EndPC:   self.readUint32(),
		}
	}
	return locVars
}

/*
	Return n byte from data.
*/
func (self *reader) readUpvalueNames() []string {
	names := make([]string, self.readUint32())
	for i := range names {
		names[i] = self.readString()
	}
	return names
}
