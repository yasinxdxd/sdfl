package sdfl

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SeqType int

const (
	SEQ_TYPE_CALL SeqType = iota
	SEQ_TYPE_FUNDEF
	SEQ_TYPE_ARG
	SEQ_TYPE_VAL
	SEQ_TYPE_LIT
	SEQ_TYPE_LEFT
	SEQ_TYPE_RIGHT
)

func seqTypeToString(s SeqType) string {
	switch s {
	case SEQ_TYPE_CALL:
		return "SEQ_TYPE_CALL"
	case SEQ_TYPE_FUNDEF:
		return "SEQ_TYPE_FUNDEF"
	case SEQ_TYPE_ARG:
		return "SEQ_TYPE_ARG"
	case SEQ_TYPE_VAL:
		return "SEQ_TYPE_VAL"
	case SEQ_TYPE_LIT:
		return "SEQ_TYPE_LIT"
	case SEQ_TYPE_LEFT:
		return "SEQ_TYPE_LEFT"
	case SEQ_TYPE_RIGHT:
		return "SEQ_TYPE_RIGHT"
	default:
		return fmt.Sprintf("Unknown: SeqType(%d)", int(s))
	}
}

type StackSeqObject struct {
	SeqType  SeqType
	RuleType *RuleType
	Id       *string
	BinopOp  *string
	LitValue *string
	Arity    int
}

func (s StackSeqObject) String() string {
	str := "{\n"

	str += "    SeqType: " + seqTypeToString(s.SeqType)
	str += "\n"
	if s.RuleType != nil {
		str += "    RuleType: " + ruleTypeToString(*s.RuleType)
		str += "\n"
	}
	if s.Id != nil {
		str += "    Id: " + *s.Id
		str += "\n"
	}
	if s.BinopOp != nil {
		str += "    BinopOp: " + *s.BinopOp
		str += "\n"
	}
	if s.LitValue != nil {
		str += "    LitValue: " + *s.LitValue
		str += "\n"
	}
	if s.Arity != -1 {
		str += "    Arity: " + strconv.FormatInt(int64(s.Arity), 10)
		str += "\n"
	}
	str += "}"

	return str
}

type Stack struct {
	Objects []StackSeqObject
}

var _stack Stack

func (s *Stack) Print() {
	for _, item := range s.Objects {
		fmt.Print(item, "\n")
	}
	fmt.Println()
}

func (s *Stack) Push(obj StackSeqObject) {
	s.Objects = append(s.Objects, obj)
}

func parseObject(objStrArr []string) StackSeqObject {
	var seqType SeqType
	var ruleType *RuleType = nil
	var id *string = nil
	var binopOp *string = nil
	var lit *string = nil
	var arity int = -1

	switch objStrArr[0] {
	case "call":
		seqType = SEQ_TYPE_CALL
		r := AST_FUN_CALL
		ruleType = &r
		i := objStrArr[1]
		id = &i
		a, _ := strconv.ParseInt(objStrArr[2], 10, 64)
		arity = int(a)
	case "fundef":
		seqType = SEQ_TYPE_FUNDEF
		r := AST_FUN_DEF
		ruleType = &r
		i := objStrArr[1]
		id = &i
		a, _ := strconv.ParseInt(objStrArr[2], 10, 64)
		arity = int(a)
	case "arg":
		seqType = SEQ_TYPE_ARG
		i := objStrArr[1]
		id = &i
		arity = 1
	case "val":
		seqType = SEQ_TYPE_VAL
		switch objStrArr[1] {
		case "number":
			r := AST_NUMBER
			ruleType = &r
			arity = 1
		case "tuple":
			r := AST_TUPLE
			ruleType = &r
			arity = 1
		case "arr":
			r := AST_ARR_EXPR
			ruleType = &r
			if objStrArr[2] == "begin" {
				a, _ := strconv.ParseInt(objStrArr[3], 10, 64)
				arity = int(a)
			}
		case "binopf":
			r := AST_BINOP_FACTOR
			ruleType = &r
			binopOp = &objStrArr[2]
			arity = 2
		case "binopt":
			r := AST_BINOP_TERM
			ruleType = &r
			binopOp = &objStrArr[2]
			arity = 2
		}

	case "literal":
		seqType = SEQ_TYPE_LIT
		lit = &objStrArr[1]
	case "left":
		seqType = SEQ_TYPE_LEFT
		arity = 1
	case "right":
		seqType = SEQ_TYPE_RIGHT
		arity = 1
	default:
		panic("unknown sequence type!")
	}

	obj := StackSeqObject{SeqType: seqType, RuleType: ruleType, Id: id, BinopOp: binopOp, LitValue: lit, Arity: arity}

	return obj
}

func Seq2AST() Program {
	// process the stack from beginning to end (index 0 to len-1)

	pos := 0
	var stmts []Stmt
	var mainExpr Expr

	// Parse all top-level items (function definitions and main expression)
	for pos < len(_stack.Objects) {
		seq := _stack.Objects[pos]

		if seq.SeqType == SEQ_TYPE_FUNDEF {
			// Parse function definition
			pos++
			stmt := parseFunctionDefinition(seq, &pos)
			stmts = append(stmts, stmt)
		} else {
			// Parse main expression (should be the scene call)
			mainExpr = parseExpression(&pos)
			break // Main expression should be the last thing
		}
	}

	return Program{
		Type:  AST_PROGRAM,
		Stmts: stmts,
		Expr:  mainExpr,
	}
}

func parseExpression(pos *int) Expr {
	if *pos >= len(_stack.Objects) {
		panic("Unexpected end of sequence")
	}

	seq := _stack.Objects[*pos]
	*pos++

	switch seq.SeqType {
	case SEQ_TYPE_CALL:
		return parseFunctionCall(seq, pos)
	case SEQ_TYPE_VAL:
		return parseValue(seq, pos)
	default:
		panic(fmt.Sprintf("Expected call or val, got: %v", seq.SeqType))
	}
}

func parseFunctionDefinition(fundefSeq StackSeqObject, pos *int) Stmt {
	// Parse function arguments (if any)
	argNames := make([]string, fundefSeq.Arity)
	for i := 0; i < fundefSeq.Arity; i++ {
		if *pos >= len(_stack.Objects) {
			panic("Expected function argument")
		}

		argSeq := _stack.Objects[*pos]
		if argSeq.SeqType != SEQ_TYPE_ARG {
			panic(fmt.Sprintf("Expected ARG for function parameter, got: %v", argSeq.SeqType))
		}

		argNames[i] = *argSeq.Id
		*pos++
	}

	// Parse function body expression
	bodyExpr := parseExpression(pos)

	funDef := &FunDef{
		Type:           AST_FUN_DEF,
		Id:             *fundefSeq.Id,
		SymbolType:     FUN_USER_DEFINED,
		FunDefArgNames: argNames,
		Expr:           &bodyExpr,
	}

	return Stmt{
		Type:   AST_FUN_DEF,
		FunDef: funDef,
	}
}

func parseFunctionCall(callSeq StackSeqObject, pos *int) Expr {
	funCall := &FunCall{
		Id:           *callSeq.Id,
		FunNamedArgs: make(map[string]FunNamedArg),
	}

	// Parse the specified number of arguments
	for i := 0; i < callSeq.Arity; i++ {
		// Next should be an ARG
		if *pos >= len(_stack.Objects) {
			panic("Expected argument in function call")
		}

		argSeq := _stack.Objects[*pos]
		if argSeq.SeqType != SEQ_TYPE_ARG {
			panic(fmt.Sprintf("Expected ARG, got: %v at position %d", argSeq.SeqType, *pos))
		}

		argName := *argSeq.Id
		*pos++

		// Parse the argument value expression
		argExpr := parseExpression(pos)

		funCall.FunNamedArgs[argName] = FunNamedArg{
			ArgName: argName,
			Expr:    argExpr,
		}
	}

	return Expr{
		Type:    AST_FUN_CALL,
		FunCall: funCall,
	}
}

func parseValue(valSeq StackSeqObject, pos *int) Expr {
	if valSeq.RuleType == nil {
		panic("Value sequence missing rule type")
	}

	switch *valSeq.RuleType {
	case AST_NUMBER:
		return parseNumberValue(pos)

	case AST_TUPLE:
		return parseTupleValue(pos)

	case AST_ARR_EXPR:
		return parseArrayValue(valSeq, pos)

	case AST_BINOP_TERM:
		return parseBinaryTerm(valSeq, pos)

	case AST_BINOP_FACTOR:
		return parseBinaryFactor(valSeq, pos)

	default:
		panic(fmt.Sprintf("Unknown value rule type: %v", *valSeq.RuleType))
	}
}

func parseNumberValue(pos *int) Expr {
	// Next should be a literal
	if *pos >= len(_stack.Objects) {
		panic("Expected literal for number")
	}

	litSeq := _stack.Objects[*pos]
	if litSeq.SeqType != SEQ_TYPE_LIT {
		panic(fmt.Sprintf("Expected literal for number, got: %v", litSeq.SeqType))
	}
	*pos++

	return Expr{
		Type:   AST_NUMBER,
		Number: &Number{Value: *litSeq.LitValue},
	}
}

func parseTupleValue(pos *int) Expr {
	// Next should be a literal
	if *pos >= len(_stack.Objects) {
		panic("Expected literal for tuple")
	}

	litSeq := _stack.Objects[*pos]
	if litSeq.SeqType != SEQ_TYPE_LIT {
		panic(fmt.Sprintf("Expected literal for tuple, got: %v", litSeq.SeqType))
	}
	*pos++

	// Parse tuple string like "(0, 0, 0)"
	tupleStr := strings.Trim(*litSeq.LitValue, "()")
	values := strings.Split(tupleStr, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	return Expr{
		Type:  AST_TUPLE,
		Tuple: &Tuple{Values: values},
	}
}

func parseArrayValue(arrSeq StackSeqObject, pos *int) Expr {
	exprs := make([]Expr, 0)

	// Parse the specified number of array elements
	for i := 0; i < arrSeq.Arity; i++ {
		expr := parseExpression(pos)
		exprs = append(exprs, expr)
	}

	// Skip "val:arr:end" if it exists
	if *pos < len(_stack.Objects) {
		nextSeq := _stack.Objects[*pos]
		if nextSeq.SeqType == SEQ_TYPE_VAL && nextSeq.RuleType != nil &&
			*nextSeq.RuleType == AST_ARR_EXPR {
			// This might be the "end" marker, skip it
			*pos++
		}
	}

	return Expr{
		Type:    AST_ARR_EXPR,
		ArrExpr: &ArrExpr{Exprs: exprs},
	}
}

func parseBinaryTerm(binopSeq StackSeqObject, pos *int) Expr {
	var leftExpr, rightExpr Expr

	// Look for LEFT marker and parse left expression
	if *pos < len(_stack.Objects) && _stack.Objects[*pos].SeqType == SEQ_TYPE_LEFT {
		*pos++ // Skip LEFT marker
		leftExpr = parseExpression(pos)
	} else {
		panic("Expected LEFT marker for binary term")
	}

	// Look for RIGHT marker and parse right expression
	if *pos < len(_stack.Objects) && _stack.Objects[*pos].SeqType == SEQ_TYPE_RIGHT {
		*pos++ // Skip RIGHT marker
		rightExpr = parseExpression(pos)
	} else {
		panic("Expected RIGHT marker for binary term")
	}

	return Expr{
		Type: AST_BINOP_TERM,
		BinopTerm: &BinopTerm{
			Left:     leftExpr,
			Right:    rightExpr,
			Operator: *binopSeq.BinopOp,
		},
	}
}

func parseBinaryFactor(binopSeq StackSeqObject, pos *int) Expr {
	var leftExpr, rightExpr Expr

	// Look for LEFT marker and parse left expression
	if *pos < len(_stack.Objects) && _stack.Objects[*pos].SeqType == SEQ_TYPE_LEFT {
		*pos++ // Skip LEFT marker
		leftExpr = parseExpression(pos)
	} else {
		panic("Expected LEFT marker for binary factor")
	}

	// Look for RIGHT marker and parse right expression
	if *pos < len(_stack.Objects) && _stack.Objects[*pos].SeqType == SEQ_TYPE_RIGHT {
		*pos++ // Skip RIGHT marker
		rightExpr = parseExpression(pos)
	} else {
		panic("Expected RIGHT marker for binary factor")
	}

	return Expr{
		Type: AST_BINOP_FACTOR,
		BinopFactor: &BinopFactor{
			Left:     leftExpr,
			Right:    rightExpr,
			Operator: *binopSeq.BinopOp,
		},
	}
}

func ParseSeq(filepath string) {
	readFile, err := os.Open(filepath)

	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		split := strings.Split(line, ":")
		stackObj := parseObject(split)
		_stack.Push(stackObj)
	}

	_stack.Print()
	readFile.Close()
}
