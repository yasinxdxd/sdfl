package sdfl

// parser

import (
	"fmt"
)

type Parser struct {
	Tokens    []Token
	token_idx int
	err       bool
}

func NewParser(tokens []Token) Parser {
	return Parser{token_idx: 0, Tokens: tokens, err: false}
}

func (p *Parser) current() Token {
	if p.token_idx >= len(p.Tokens) {
		return Token{Kind: EOF}
	}
	return p.Tokens[p.token_idx]
}

func (p *Parser) lookAhead(num int) Token {
	idx := p.token_idx + num
	if idx >= len(p.Tokens) {
		return Token{Kind: EOF}
	} else if idx <= 0 {
		panic("PANIC: You go too back far there is no token!\n")
	}
	return p.Tokens[idx]
}

func (p *Parser) eat(token_kind TokenType) (bool, Token) {
	tok := p.current()
	if token_kind != tok.Kind {
		fmt.Printf("ERROR:%d:%d expected %s but got %s\n", tok.Row, tok.Col, TokenName[token_kind], TokenName[tok.Kind])
		p.err = true
		return false, tok
	}
	p.token_idx++
	return true, tok
}

func (p *Parser) eatOneOf(kinds ...TokenType) (bool, Token) {
	tok := p.current()
	for _, kind := range kinds {
		if tok.Kind == kind {
			p.token_idx++
			return true, tok
		}
	}
	// Build expected kinds list for error message
	expected := ""
	for i, k := range kinds {
		if i > 0 {
			expected += " or "
		}
		expected += TokenName[k]
	}
	fmt.Printf("ERROR:%d:%d expected %s but got %s\n", tok.Row, tok.Col, expected, TokenName[tok.Kind])
	p.err = true
	return false, tok
}

func (p *Parser) IsThereError() bool {
	return p.err
}

func (p *Parser) isBinOp() bool {
	switch p.current().Kind {
	case PUNC_PLUS, PUNC_SUB, PUNC_MULT, PUNC_DIV:
		return true
	default:
		return false
	}
}

func (p *Parser) getOpPrecedence(op TokenType) int {
	switch op {
	case PUNC_MULT, PUNC_DIV:
		return 3
	case PUNC_PLUS, PUNC_SUB:
		return 2
	default:
		return 0
	}
}

func (p *Parser) ParseFunDef() FunDef {
	p.eat(KW_DEF)
	_, tok := p.eat(KW_ID)
	funName := tok.Value
	p.eat(PUNC_LPAREN)
	funDefArgNames := []string{}
	for p.current().Kind != PUNC_RPAREN {
		_, tok := p.eat(KW_ID)
		funDefArgNames = append(funDefArgNames, tok.Value)
		if p.current().Kind != PUNC_RPAREN {
			p.eat(PUNC_COMMA)
		}
	}
	p.eat(PUNC_RPAREN)
	p.eat(PUNC_LCURLY)
	expr := p.ParseExpr()
	p.eat(PUNC_RCURLY)

	funDef := FunDef{Type: AST_FUN_DEF, SymbolType: FUN_USER_DEFINED, Id: funName, FunDefArgNames: funDefArgNames, Expr: &expr}
	functionSymbols[funName] = funDef
	return funDef
}

func (p *Parser) ParseFunNamedArg() FunNamedArg {
	_, tok := p.eat(KW_ID)
	argName := tok.Value

	p.eat(PUNC_COLON)
	expr := p.ParseExpr()

	funNamedArg := FunNamedArg{ArgName: argName, Expr: expr}
	return funNamedArg
}

func (p *Parser) ParseFunCall() FunCall {
	_, tok := p.eat(KW_ID)
	p.eat(PUNC_LPAREN)

	funNamedArgs := map[string]FunNamedArg{}
	for p.current().Kind != PUNC_RPAREN {
		funNamedArg := p.ParseFunNamedArg()
		funNamedArgs[funNamedArg.ArgName] = funNamedArg
		if p.current().Kind != PUNC_RPAREN {
			p.eat(PUNC_COMMA)
		}
	}

	p.eat(PUNC_RPAREN)
	funcCall := FunCall{Id: tok.Value, FunNamedArgs: funNamedArgs}
	return funcCall
}

func (p *Parser) ParseNumber() Number {
	_, tok := p.eat(NUMBER_FLOAT)
	number := Number{Value: tok.Value}
	return number
}

func (p *Parser) ParseTuple() Tuple {
	p.eat(PUNC_LPAREN)

	values := []string{}
	for p.current().Kind == NUMBER_FLOAT {
		_, tok := p.eat(NUMBER_FLOAT)
		values = append(values, tok.Value)
		if p.current().Kind != PUNC_RPAREN {
			p.eat(PUNC_COMMA)
		}
	}

	p.eat(PUNC_RPAREN)

	tuple := Tuple{Values: values}
	return tuple
}

func (p *Parser) ParseArrExpr() ArrExpr {
	p.eat(PUNC_LSQUARE)
	exprs := []Expr{}
	for p.current().Kind != PUNC_RSQUARE {
		expr := p.ParseExpr()
		exprs = append(exprs, expr)
		if p.current().Kind != PUNC_RSQUARE {
			p.eat(PUNC_COMMA)
		}
	}
	p.eat(PUNC_RSQUARE)

	arrExpr := ArrExpr{Exprs: exprs}
	return arrExpr
}

func (p *Parser) ParseExpr() Expr {
	expr := Expr{}
	if p.current().Kind == NUMBER_FLOAT {
		number := p.ParseNumber()
		expr.Number = &number
		expr.Type = AST_NUMBER
	} else if p.current().Kind == KW_ID {
		if p.lookAhead(1).Kind == PUNC_LPAREN {
			funcCall := p.ParseFunCall()
			expr.Type = AST_FUN_CALL
			expr.FunCall = &funcCall
		} else {
			// TODO: check a symbol table if a variable created
		}
	} else if p.current().Kind == PUNC_LPAREN {
		if p.lookAhead(1).Kind == NUMBER_FLOAT {
			tuple := p.ParseTuple()
			expr.Type = RuleType(int(AST_TUPLE) + len(tuple.Values))
			expr.Tuple = &tuple
		}
	} else if p.current().Kind == PUNC_LSQUARE {
		arrExpr := p.ParseArrExpr()
		expr.ArrExpr = &arrExpr
		expr.Type = AST_ARR_EXPR
	}

	return expr
}

func (p *Parser) ParseStmt() Stmt {
	stmt := Stmt{}
	if p.current().Kind == KW_DEF {
		stmt.Type = AST_FUN_DEF
		fundef := p.ParseFunDef()
		stmt.FunDef = &fundef
	}

	return stmt
}

func (p *Parser) Parse() Program {
	stmts := []Stmt{}
	for p.current().Kind == KW_DEF || p.current().Kind == KW_LET {
		stmt := p.ParseStmt()
		stmts = append(stmts, stmt)
	}

	expr := p.ParseExpr()

	if !p.err {
		println("DONE!")
	}

	program := Program{Type: AST_PROGRAM, Expr: expr, Stmts: stmts}
	return program
}
