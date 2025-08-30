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
	AST_TUPLE1
	AST_TUPLE2
	AST_TUPLE3
	AST_NUMBER
	AST_ARR_EXPR
)

type Program struct {
	Type  RuleType
	Exprs []Expr
}

type Expr struct {
	Type    RuleType
	FunCall *FunCall
	Tuple   *Tuple
	ArrExpr *ArrExpr
	Number  *Number
}

type Number struct {
	Value string
}

type Tuple struct {
	Values []string
}

type SymbolType int

const (
	FUN_BUILTIN_SCENE SymbolType = iota
	FUN_BUILTIN_CAMERA
	FUN_BUILTIN
	FUN_USER_DEFINED
	VAR_BUILTIN
	VAR_USER_DEFINED
)

type FunDef struct {
	Type           RuleType
	SymbolType     SymbolType
	Id             string
	FunDefArgNames []string
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
	for _, expr := range prog.Exprs {
		printExpr(expr, level+1)
	}
}

func printExpr(expr Expr, level int) {
	switch expr.Type {
	case AST_FUN_CALL:
		printFunCall(expr.FunCall, level)
	case AST_TUPLE1:
		fallthrough
	case AST_TUPLE2:
		fallthrough
	case AST_TUPLE3:
		printTuple(expr.Tuple, level)
	case AST_ARR_EXPR:
		printArr(expr.ArrExpr, level)
	case AST_NUMBER:
		printNumber(expr.Number, level)
	default:
		fmt.Printf("%sUnknown expr type: %v\n", indent(level), expr.Type)
	}
}

func printNumber(num *Number, level int) {
	fmt.Printf("%sNumber: %s\n", indent(level), num.Value)
}

func printTuple(tuple *Tuple, level int) {
	fmt.Printf("%sTuple:\n", indent(level))
	for _, val := range tuple.Values {
		fmt.Printf("%s- %s\n", indent(level+1), val)
	}
}

func printFunCall(fun *FunCall, level int) {
	fmt.Printf("%sFunCall: %s\n", indent(level), fun.Id)
	for _, arg := range fun.FunNamedArgs {
		fmt.Printf("%sArg: %s\n", indent(level+1), arg.ArgName)
		printExpr(arg.Expr, level+2)
	}
}

func printArr(arr *ArrExpr, level int) {
	fmt.Printf("%sArray:\n", indent(level))
	for _, e := range arr.Exprs {
		printExpr(e, level+1)
	}
}
