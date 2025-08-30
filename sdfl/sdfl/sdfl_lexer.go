package sdfl

import (
	"fmt"
	"regexp"
)

type TokenType int

const (
	EOF    TokenType = -1
	KW_LET TokenType = iota
	KW_ID
	NUMBER_FLOAT
	NUMBER_INT
	PUNC_MULT
	PUNC_DIV
	PUNC_PLUS
	PUNC_SUB
	PUNC_EQUAL
	PUNC_LPAREN
	PUNC_RPAREN
	PUNC_LSQUARE
	PUNC_RSQUARE
	PUNC_COLON
	PUNC_COMMA
	WS
)

var TokenName = map[TokenType]string{
	EOF:          "EOF",
	KW_LET:       "KW_LET",
	KW_ID:        "KW_ID",
	NUMBER_FLOAT: "NUMBER_FLOAT",
	NUMBER_INT:   "NUMBER_INT",
	PUNC_MULT:    "PUNC_MULT",
	PUNC_DIV:     "PUNC_DIV",
	PUNC_PLUS:    "PUNC_PLUS",
	PUNC_SUB:     "PUNC_SUB",
	PUNC_EQUAL:   "PUNC_EQUAL",
	PUNC_LPAREN:  "PUNC_LPAREN",
	PUNC_RPAREN:  "PUNC_RPAREN",
	PUNC_LSQUARE: "PUNC_LSQUARE",
	PUNC_RSQUARE: "PUNC_RSQUARE",
	PUNC_COLON:   "PUNC_COLON",
	PUNC_COMMA:   "PUNC_COMMA",
	WS:           "WS",
}

type Token struct {
	Kind  TokenType
	Value string
	Row   int
	Col   int
}

type Rule struct {
	kind     TokenType
	regex    regexp.Regexp
	skipable bool
}

var rules = []Rule{}

func InitRules() {
	reg_KW_LET := regexp.MustCompile(`let`)
	if reg_KW_LET != nil {
		rules = append(rules, Rule{kind: KW_LET, regex: *reg_KW_LET, skipable: false})
	}
	reg_KW_ID := regexp.MustCompile(`[a-zA-Z_][a-zA-Z_0-9]*`)
	if reg_KW_ID != nil {
		rules = append(rules, Rule{kind: KW_ID, regex: *reg_KW_ID, skipable: false})
	}
	reg_NUMBER_FLOAT := regexp.MustCompile(`[+-]?(?:\d+\.\d*|\.\d+|\d+)(?:[eE][+-]?\d+)?[fF]?`)
	if reg_NUMBER_FLOAT != nil {
		rules = append(rules, Rule{kind: NUMBER_FLOAT, regex: *reg_NUMBER_FLOAT, skipable: false})
	}
	reg_NUMBER_INT := regexp.MustCompile(`[+-]?(?:0[xX][0-9A-Fa-f]+|\d+)[uU]?`)
	if reg_NUMBER_INT != nil {
		rules = append(rules, Rule{kind: NUMBER_INT, regex: *reg_NUMBER_INT, skipable: false})
	}
	reg_PUNC_MULT := regexp.MustCompile(`[*]`)
	if reg_PUNC_MULT != nil {
		rules = append(rules, Rule{kind: PUNC_MULT, regex: *reg_PUNC_MULT, skipable: false})
	}
	reg_PUNC_DIV := regexp.MustCompile(`[/]`)
	if reg_PUNC_DIV != nil {
		rules = append(rules, Rule{kind: PUNC_DIV, regex: *reg_PUNC_DIV, skipable: false})
	}
	reg_PUNC_PLUS := regexp.MustCompile(`[+]`)
	if reg_PUNC_PLUS != nil {
		rules = append(rules, Rule{kind: PUNC_PLUS, regex: *reg_PUNC_PLUS, skipable: false})
	}
	reg_PUNC_SUB := regexp.MustCompile(`[-]`)
	if reg_PUNC_SUB != nil {
		rules = append(rules, Rule{kind: PUNC_SUB, regex: *reg_PUNC_SUB, skipable: false})
	}
	reg_PUNC_EQUAL := regexp.MustCompile(`=`)
	if reg_PUNC_EQUAL != nil {
		rules = append(rules, Rule{kind: PUNC_EQUAL, regex: *reg_PUNC_EQUAL, skipable: false})
	}
	reg_PUNC_LPAREN := regexp.MustCompile(`[(]`)
	if reg_PUNC_LPAREN != nil {
		rules = append(rules, Rule{kind: PUNC_LPAREN, regex: *reg_PUNC_LPAREN, skipable: false})
	}
	reg_PUNC_RPAREN := regexp.MustCompile(`[)]`)
	if reg_PUNC_RPAREN != nil {
		rules = append(rules, Rule{kind: PUNC_RPAREN, regex: *reg_PUNC_RPAREN, skipable: false})
	}
	reg_PUNC_LSQUARE := regexp.MustCompile(`[[]`)
	if reg_PUNC_LSQUARE != nil {
		rules = append(rules, Rule{kind: PUNC_LSQUARE, regex: *reg_PUNC_LSQUARE, skipable: false})
	}
	reg_PUNC_RSQUARE := regexp.MustCompile(`[]]`)
	if reg_PUNC_RSQUARE != nil {
		rules = append(rules, Rule{kind: PUNC_RSQUARE, regex: *reg_PUNC_RSQUARE, skipable: false})
	}
	reg_PUNC_COLON := regexp.MustCompile(`:`)
	if reg_PUNC_COLON != nil {
		rules = append(rules, Rule{kind: PUNC_COLON, regex: *reg_PUNC_COLON, skipable: false})
	}
	reg_PUNC_COMMA := regexp.MustCompile(`,`)
	if reg_PUNC_COMMA != nil {
		rules = append(rules, Rule{kind: PUNC_COMMA, regex: *reg_PUNC_COMMA, skipable: false})
	}
	reg_WS := regexp.MustCompile(`[ \t\r\n]`)
	if reg_WS != nil {
		rules = append(rules, Rule{kind: WS, regex: *reg_WS, skipable: true})
	}
}

func Tokenize(input string) []Token {
	pos := 0
	inputLen := len(input)
	tokens := []Token{}
	row := 1
	col := 1
	for pos < inputLen {
		matched := false
		skip := false
		if input[pos] == '\n' {
			row++
			col = 0
		}
		for _, rule := range rules {
			loc := rule.regex.FindStringIndex(input[pos:])
			if loc != nil && loc[0] == 0 {
				// determine token type based on index in regexes
				skip = rule.skipable
				tokenType := rule.kind
				value := input[pos : pos+loc[1]]
				if !skip {
					tokens = append(tokens, Token{Kind: tokenType, Value: value, Row: row, Col: col})
				}
				col += loc[1]
				pos += loc[1]
				matched = true
				break
			}
		}
		if !matched {
			fmt.Println("ERROR: unrecognized token occured!")
			pos++
			col++
		}
	}

	tokens = append(tokens, Token{Kind: EOF, Value: "EOF"})
	return tokens
}
