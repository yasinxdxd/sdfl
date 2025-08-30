package main

import (
	"fmt"
	"os"

	"./sdfl"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		println("ERROR: sflc <input.sdfl>")
		return
	}
	filePath := args[0]

	sdfl.InitRules()
	source, err := os.ReadFile(filePath)
	check(err)
	fmt.Println(string(source))

	tokens := sdfl.Tokenize(string(source))

	for _, t := range tokens {
		fmt.Printf("Token:%d:%d %-10s Value: %q\n", t.Row, t.Col, sdfl.TokenName[t.Kind], t.Value)
	}
	parser := sdfl.NewParser(tokens)
	program := parser.Parse()

	sdfl.PrintAST(program)
	sdfl.Generate(&program)

	// fmt.Println(sdfl.GetCode())
	f, err := os.Create("out_frag.glsl")
	check(err)
	defer f.Close()
	len, err := f.WriteString(sdfl.GetCode())
	check(err)
	fmt.Println(len, "bytes written successfully")
}
