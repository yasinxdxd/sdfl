package sdfl

import (
	"fmt"
	"strings"
)

// ast
type RuleType int

const (
	AST_PROGRAM RuleType = iota
	AST_FUN_DEF
	AST_FUN_CALL
	AST_TUPLE
	AST_NUMBER
	AST_BINOP_TERM
	AST_BINOP_FACTOR
	AST_ARR_EXPR
)

func ruleTypeToString(r RuleType) string {
	switch r {
	case AST_PROGRAM:
		return "AST_PROGRAM"
	case AST_FUN_DEF:
		return "AST_FUN_DEF"
	case AST_FUN_CALL:
		return "AST_FUN_CALL"
	case AST_TUPLE:
		return "AST_TUPLE"
	case AST_NUMBER:
		return "AST_NUMBER"
	case AST_BINOP_TERM:
		return "AST_BINOP_TERM"
	case AST_BINOP_FACTOR:
		return "AST_BINOP_FACTOR"
	case AST_ARR_EXPR:
		return "AST_ARR_EXPR"
	default:
		return fmt.Sprintf("Unknown: RuleType(%d)", int(r))
	}
}

type Program struct {
	Type  RuleType
	Stmts []Stmt
	Expr  Expr
}

type Stmt struct {
	Type   RuleType
	FunDef *FunDef
}

type Expr struct {
	Type           RuleType
	FunCall        *FunCall
	Tuple          *Tuple
	ArrExpr        *ArrExpr
	Number         *Number
	BinopTerm      *BinopTerm
	BinopFactor    *BinopFactor
	HasParentheses bool
}

type Number struct {
	Value string
}

type Tuple struct {
	Values []string
}

type BinopTerm struct {
	Left     Expr
	Right    Expr
	Operator string
}

type BinopFactor struct {
	Left     Expr
	Right    Expr
	Operator string
}

type SymbolType int

const (
	FUN_BUILTIN_SCENE SymbolType = iota
	FUN_BUILTIN_LOCAL
	FUN_BUILTIN_CAMERA
	FUN_BUILTIN_ROTATE_AROUND
	FUN_BUILTIN_OP
	FUN_BUILTIN_SHAPE
	FUN_BUILTIN_SDFL
	FUN_BUILTIN_GLSL
	FUN_USER_DEFINED
	VAR_BUILTIN
	VAR_USER_DEFINED
)

type FunDef struct {
	Type           RuleType
	SymbolType     SymbolType
	Id             string
	FunDefArgNames []string
	Expr           *Expr
}

type FunNamedArg struct {
	ArgName string
	Expr    Expr
}

type FunCall struct {
	Id           string
	FunNamedArgs map[string]FunNamedArg
}

type ArrExpr struct {
	Exprs []Expr
}

/*
	AST Print
*/

func PrintAST(prog Program) {
	printProgram(prog, 0)
}

func indent(level int) string {
	return strings.Repeat("  ", level)
}

func printProgram(prog Program, level int) {
	fmt.Printf("%sProgram:\n", indent(level))
	if len(prog.Stmts) > 0 {
		fmt.Printf("%sStatements:\n", indent(level+1))
		for i, stmt := range prog.Stmts {
			fmt.Printf("%s[%d]\n", indent(level+2), i)
			printStmt(stmt, level+3)
		}
	}

	// print main expression (scene)
	fmt.Printf("%sExpression:\n", indent(level+1))
	printExpr(prog.Expr, level+2)
}

func printStmt(stmt Stmt, level int) {
	switch stmt.Type {
	case AST_FUN_DEF:
		printFunDef(stmt.FunDef, level)
	default:
		fmt.Printf("%sUnknown statement type: %v\n", indent(level), stmt.Type)
	}
}

func printFunDef(funDef *FunDef, level int) {
	fmt.Printf("%sFunDef:\n", indent(level))
	fmt.Printf("%sId: %s\n", indent(level+1), funDef.Id)
	fmt.Printf("%sSymbolType: %s\n", indent(level+1), symbolTypeToString(funDef.SymbolType))

	if len(funDef.FunDefArgNames) > 0 {
		fmt.Printf("%sArguments:\n", indent(level+1))
		for i, argName := range funDef.FunDefArgNames {
			fmt.Printf("%s[%d] %s\n", indent(level+2), i, argName)
		}
	}

	if funDef.Expr != nil {
		fmt.Printf("%sBody:\n", indent(level+1))
		printExpr(*funDef.Expr, level+2)
	}
}

func symbolTypeToString(st SymbolType) string {
	switch st {
	case FUN_BUILTIN_SCENE:
		return "FUN_BUILTIN_SCENE"
	case FUN_BUILTIN_CAMERA:
		return "FUN_BUILTIN_CAMERA"
	case FUN_BUILTIN_ROTATE_AROUND:
		return "FUN_BUILTIN_ROTATE_AROUND"
	case FUN_BUILTIN_OP:
		return "FUN_BUILTIN_OP"
	case FUN_BUILTIN_SHAPE:
		return "FUN_BUILTIN_SHAPE"
	case FUN_BUILTIN_SDFL:
		return "FUN_BUILTIN_SDFL"
	case FUN_BUILTIN_GLSL:
		return "FUN_BUILTIN_GLSL"
	case FUN_USER_DEFINED:
		return "FUN_USER_DEFINED"
	case VAR_BUILTIN:
		return "VAR_BUILTIN"
	case VAR_USER_DEFINED:
		return "VAR_USER_DEFINED"
	default:
		return "UNKNOWN_SYMBOL_TYPE"
	}
}

func printExpr(expr Expr, level int) {
	switch expr.Type {
	case AST_FUN_CALL:
		printFunCall(expr.FunCall, level)
	case AST_TUPLE:
		printTuple(expr.Tuple, level)
	case AST_ARR_EXPR:
		printArr(expr.ArrExpr, level)
	case AST_NUMBER:
		printNumber(expr.Number, level)
	case AST_BINOP_TERM:
		printBinopTerm(expr.BinopTerm, level)
	case AST_BINOP_FACTOR:
		printBinopFactor(expr.BinopFactor, level)
	default:
		fmt.Printf("%sUnknown expr type: %v\n", indent(level), expr.Type)
	}
}

func printNumber(num *Number, level int) {
	fmt.Printf("%sNumber: %s\n", indent(level), num.Value)
}

func printTuple(tuple *Tuple, level int) {
	fmt.Printf("%sTuple:\n", indent(level))
	for i, val := range tuple.Values {
		fmt.Printf("%s[%d] %s\n", indent(level+1), i, val)
	}
}

func printFunCall(fun *FunCall, level int) {
	fmt.Printf("%sFunCall: %s\n", indent(level), fun.Id)
	if len(fun.FunNamedArgs) > 0 {
		fmt.Printf("%sArguments:\n", indent(level+1))
		for argName, arg := range fun.FunNamedArgs {
			fmt.Printf("%s%s:\n", indent(level+2), argName)
			printExpr(arg.Expr, level+3)
		}
	}
}

func printArr(arr *ArrExpr, level int) {
	fmt.Printf("%sArray:\n", indent(level))
	for i, e := range arr.Exprs {
		fmt.Printf("%s[%d]\n", indent(level+1), i)
		printExpr(e, level+2)
	}
}

func printBinopTerm(binop *BinopTerm, level int) {
	fmt.Printf("%sBinaryOperation (Term): %s\n", indent(level), binop.Operator)
	fmt.Printf("%sLeft:\n", indent(level+1))
	printExpr(binop.Left, level+2)
	fmt.Printf("%sRight:\n", indent(level+1))
	printExpr(binop.Right, level+2)
}

func printBinopFactor(binop *BinopFactor, level int) {
	fmt.Printf("%sBinaryOperation (Factor): %s\n", indent(level), binop.Operator)
	fmt.Printf("%sLeft:\n", indent(level+1))
	printExpr(binop.Left, level+2)
	fmt.Printf("%sRight:\n", indent(level+1))
	printExpr(binop.Right, level+2)
}
