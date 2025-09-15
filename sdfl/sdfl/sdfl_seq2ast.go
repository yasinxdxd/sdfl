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

func (s *Stack) Top() StackSeqObject {
	if s.IsEmpty() {
		panic("stack is empty")
	}
	return s.Objects[len(s.Objects)-1]
}

func (s *Stack) IsEmpty() bool {
	if len(s.Objects) == 0 {
		return true
	}
	return false
}

func (s *Stack) Print() {
	for _, item := range s.Objects {
		fmt.Print(item, "\n")
	}
	fmt.Println()
}

func (s *Stack) Push(obj StackSeqObject) {
	s.Objects = append(s.Objects, obj)
}

func (s *Stack) Pop() {
	if s.IsEmpty() {
		return
	}
	s.Objects = s.Objects[:len(s.Objects)-1]
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
	// TODO: implement this
	for !_stack.IsEmpty() {
		seq := _stack.Top()

		switch seq.SeqType {
		case SEQ_TYPE_CALL:

		case SEQ_TYPE_FUNDEF:

		case SEQ_TYPE_ARG:

		case SEQ_TYPE_VAL:

		case SEQ_TYPE_LIT:

		case SEQ_TYPE_LEFT:

		case SEQ_TYPE_RIGHT:

		default:
			panic("Seq2AST:ERROR: Unknown sequence!")
		}
	}
	return Program{}
}

func ParseSeq() {
	// TODO no need for file I think
	readFile, err := os.Open("ast_sequence.txt")

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
