/*
	See Copyright Notice at LICENSE file
*/
package compiler

import (
	. "goluar/common"
	"math"
)

/*
	@description
		Compute logical 'and', 'or' in the expression.
	@param
		exp		BinopExp	"binary opertor expression"
	@return
		exp		Exp		"Exp struct is defiend in ast_exp.go"
*/
func optimizeLogicalAndOr(exp *BinopExp) Exp {
	switch exp.Op {
	case LEX_OP_AND:
		return optimizeLogicalOr(exp)
	case LEX_OP_OR:
		return optimizeLogicalAnd(exp)
	default:
		return exp
	}
}

/*
	@description
		Compute ('+' | '-' | '*' | '%' | '/') in the expression.
	@param
		exp		BinopExp	"binary opertor expression"
	@return
		exp		Exp		"Exp struct is defiend in ast_exp.go"
*/
func optimizeArithBinaryOp(exp *BinopExp) Exp {
	if x, ok := exp.Exp1.(*IntegerExp); ok {
		if y, ok := exp.Exp2.(*IntegerExp); ok {
			switch exp.Op {
			case LEX_OP_ADD:
				return &IntegerExp{exp.Line, x.Val + y.Val}
			case LEX_OP_SUB:
				return &IntegerExp{exp.Line, x.Val - y.Val}
			case LEX_OP_MUL:
				return &IntegerExp{exp.Line, x.Val * y.Val}
			case LEX_OP_MOD:
				if y.Val != 0 {
					return &IntegerExp{exp.Line, IMod(x.Val, y.Val)}
				}
			}
		}
	}
	if f, ok := castToFloat(exp.Exp1); ok {
		if g, ok := castToFloat(exp.Exp2); ok {
			switch exp.Op {
			case LEX_OP_ADD:
				return &FloatExp{exp.Line, f + g}
			case LEX_OP_SUB:
				return &FloatExp{exp.Line, f - g}
			case LEX_OP_MUL:
				return &FloatExp{exp.Line, f * g}
			case LEX_OP_DIV:
				if g != 0 {
					return &FloatExp{exp.Line, f / g}
				}
			case LEX_OP_MOD:
				if g != 0 {
					return &FloatExp{exp.Line, FMod(f, g)}
				}
			case LEX_OP_POW:
				return &FloatExp{exp.Line, math.Pow(f, g)}
			}
		}
	}
	return exp
}

/*
	@description
		Compute pow in the expression.
	@param
		exp		BinopExp	"binary opertor expression"
	@return
		exp		Exp		"Exp struct is defiend in ast_exp.go"
*/
func optimizePow(exp Exp) Exp {
	if binop, ok := exp.(*BinopExp); ok {
		if binop.Op == LEX_OP_POW {
			binop.Exp2 = optimizePow(binop.Exp2)
		}
		return optimizeArithBinaryOp(binop)
	}
	return exp
}

/*
	@description
		Compute unique operator ('-'|'not'|'~') in the expression.
	@param
		exp		UnopExp	"Unique opertor expression"
	@return
		exp		Exp		"Exp struct is defiend in ast_exp.go"
*/
func optimizeUnaryOp(exp *UnopExp) Exp {
	switch exp.Op {
	case LEX_OP_UNM:
		return optimizeUnm(exp)
	case LEX_OP_NOT:
		return optimizeNot(exp)
	default:
		return exp
	}
}

// Compute '-' in the expression.
func optimizeUnm(exp *UnopExp) Exp {
	switch x := exp.Exp.(type) {
	case *IntegerExp:
		x.Val = -x.Val
		return x
	case *FloatExp:
		if x.Val != 0 {
			x.Val = -x.Val
			return x
		}
	}
	return exp
}

// Compute 'not' in the expression.
func optimizeNot(exp *UnopExp) Exp {
	switch exp.Exp.(type) {
	case *NilExp, *FalseExp: // false
		return &TrueExp{exp.Line}
	case *TrueExp, *IntegerExp, *FloatExp, *StringExp: // true
		return &FalseExp{exp.Line}
	default:
		return exp
	}
}

// Compute '~' in the expression.
func optimizeBnot(exp *UnopExp) Exp {
	switch x := exp.Exp.(type) {
	case *IntegerExp:
		x.Val = ^x.Val
		return x
	case *FloatExp:
		if i, ok := FloatToInteger(x.Val); ok {
			return &IntegerExp{x.Line, ^i}
		}
	}
	return exp
}

// Compute logical 'or' in the expression.
func optimizeLogicalOr(exp *BinopExp) Exp {
	if isTrue(exp.Exp1) {
		return exp.Exp1 // true or x => true
	}
	if isFalse(exp.Exp1) && !isVarargsOrFuncCall(exp.Exp2) {
		return exp.Exp2 // false or x => x
	}
	return exp
}

// Compute logical 'and' in the expression.
func optimizeLogicalAnd(exp *BinopExp) Exp {
	if isFalse(exp.Exp1) {
		return exp.Exp1 // false and x => false
	}
	if isTrue(exp.Exp1) && !isVarargsOrFuncCall(exp.Exp2) {
		return exp.Exp2 // true and x => x
	}
	return exp
}

// Return whether a expression is false.
// false and nil are belong to false.
func isFalse(exp Exp) bool {
	switch exp.(type) {
	case *FalseExp, *NilExp:
		return true
	default:
		return false
	}
}

// Return whether a expression is true.
// true, integer, float, string are belong to true.
func isTrue(exp Exp) bool {
	switch exp.(type) {
	case *TrueExp, *IntegerExp, *FloatExp, *StringExp:
		return true
	default:
		return false
	}
}

// Whether is variable arguments or function calling.
func isVarargsOrFuncCall(exp Exp) bool {
	switch exp.(type) {
	case *VarargExp, *FuncCallExp:
		return true
	}
	return false
}

// Cast a expression to integer.
func castToInt(exp Exp) (int64, bool) {
	switch x := exp.(type) {
	case *IntegerExp:
		return x.Val, true
	case *FloatExp:
		return FloatToInteger(x.Val)
	default:
		return 0, false
	}
}

// Cast a expression to float.
func castToFloat(exp Exp) (float64, bool) {
	switch x := exp.(type) {
	case *IntegerExp:
		return float64(x.Val), true
	case *FloatExp:
		return x.Val, true
	default:
		return 0, false
	}
}
