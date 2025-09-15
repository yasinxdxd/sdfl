package sdfl

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ASTToSequence(prog Program) string {
	var lines []string

	for _, stmt := range prog.Stmts {
		lines = append(lines, stmtToLines(stmt)...)
	}

	lines = append(lines, exprToLines(prog.Expr)...)

	return strings.Join(lines, "\n")
}

func stmtToLines(stmt Stmt) []string {
	switch stmt.Type {
	case AST_FUN_DEF:
		return funDefToLines(stmt.FunDef)
	default:
		return []string{"unknown_stmt"}
	}
}

func funDefToLines(funDef *FunDef) []string {
	var lines []string

	lines = append(lines, "fundef:"+funDef.Id+":"+strconv.Itoa(len(funDef.FunDefArgNames)))

	// add argument names
	for _, argName := range funDef.FunDefArgNames {
		lines = append(lines, "param:"+argName)
	}

	// add body expression
	if funDef.Expr != nil {
		lines = append(lines, exprToLines(*funDef.Expr)...)
	}

	lines = append(lines, "fundef:end")
	return lines
}

func exprToLines(expr Expr) []string {
	var lines []string

	if expr.HasParentheses {
		lines = append(lines, "paren:open")
	}

	switch expr.Type {
	case AST_FUN_CALL:
		lines = append(lines, funCallToLines(expr.FunCall)...)
	case AST_TUPLE:
		lines = append(lines, tupleToLines(expr.Tuple)...)
	case AST_ARR_EXPR:
		lines = append(lines, arrExprToLines(expr.ArrExpr)...)
	case AST_NUMBER:
		lines = append(lines, numberToLines(expr.Number)...)
	case AST_BINOP_TERM:
		lines = append(lines, binopTermToLines(expr.BinopTerm)...)
	case AST_BINOP_FACTOR:
		lines = append(lines, binopFactorToLines(expr.BinopFactor)...)
	default:
		lines = append(lines, "unknown_expr")
	}

	if expr.HasParentheses {
		lines = append(lines, "paren:close")
	}

	return lines
}

func funCallToLines(funCall *FunCall) []string {
	var lines []string

	lines = append(lines, "call:"+funCall.Id+":"+strconv.Itoa(len(funCall.FunNamedArgs)))

	// process each argument in a consistent order (sorted by name for reproducibility)
	if len(funCall.FunNamedArgs) > 0 {
		argNames := make([]string, 0, len(funCall.FunNamedArgs))
		for argName := range funCall.FunNamedArgs {
			argNames = append(argNames, argName)
		}

		// sort argument names for consistent ordering
		for i := 0; i < len(argNames); i++ {
			for j := i + 1; j < len(argNames); j++ {
				if argNames[i] > argNames[j] {
					argNames[i], argNames[j] = argNames[j], argNames[i]
				}
			}
		}

		for _, argName := range argNames {
			arg := funCall.FunNamedArgs[argName]
			lines = append(lines, "arg:"+argName)
			lines = append(lines, exprToLines(arg.Expr)...)
		}
	}

	return lines
}

func tupleToLines(tuple *Tuple) []string {
	var lines []string

	lines = append(lines, "val:tuple")

	// Add the actual tuple values as a formatted line
	if len(tuple.Values) > 0 {
		tupleStr := "(" + strings.Join(tuple.Values, ", ") + ")"
		lines = append(lines, "literal:"+tupleStr)
	}

	return lines
}

func arrExprToLines(arrExpr *ArrExpr) []string {
	var lines []string

	lines = append(lines, "val:arr:begin")

	for _, expr := range arrExpr.Exprs {
		lines = append(lines, exprToLines(expr)...)
	}

	lines = append(lines, "val:arr:end")
	return lines
}

func numberToLines(number *Number) []string {
	var lines []string

	lines = append(lines, "val:number")
	lines = append(lines, "literal:"+number.Value)

	return lines
}

// binopTermToLines converts a binary operation (term) to lines
func binopTermToLines(binop *BinopTerm) []string {
	var lines []string

	lines = append(lines, "val:binopt:"+binop.Operator)
	lines = append(lines, "left")
	lines = append(lines, exprToLines(binop.Left)...)
	lines = append(lines, "right")
	lines = append(lines, exprToLines(binop.Right)...)

	return lines
}

// binopFactorToLines converts a binary operation (factor) to lines
func binopFactorToLines(binop *BinopFactor) []string {
	var lines []string

	lines = append(lines, "val:binopf:"+binop.Operator)
	lines = append(lines, "left")
	lines = append(lines, exprToLines(binop.Left)...)
	lines = append(lines, "right")
	lines = append(lines, exprToLines(binop.Right)...)

	return lines
}

// WriteSequenceToFile writes the AST sequence to a file
func WriteSequenceToFile(prog Program, filename string, sequence string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	_, err = file.WriteString(sequence)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	return nil
}

// WriteMultipleSequencesToFile writes multiple AST sequences to a file, separated by empty lines
func WriteMultipleSequencesToFile(programs []Program, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	for i, prog := range programs {
		sequence := ASTToSequence(prog)
		_, err = file.WriteString(sequence)
		if err != nil {
			return fmt.Errorf("failed to write sequence %d to file %s: %w", i, filename, err)
		}

		// Add double newline between sequences if not the last one
		if i < len(programs)-1 {
			_, err = file.WriteString("\n\n")
			if err != nil {
				return fmt.Errorf("failed to write separator to file %s: %w", filename, err)
			}
		}
	}

	return nil
}

// AppendSequenceToFile appends an AST sequence to an existing file
func AppendSequenceToFile(prog Program, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	// Check if file is not empty, add separator before appending
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", filename, err)
	}

	sequence := ASTToSequence(prog)

	if fileInfo.Size() > 0 {
		_, err = file.WriteString("\n\n" + sequence)
	} else {
		_, err = file.WriteString(sequence)
	}

	if err != nil {
		return fmt.Errorf("failed to append to file %s: %w", filename, err)
	}

	return nil
}
