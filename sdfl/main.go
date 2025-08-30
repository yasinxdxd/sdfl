package main

import (
	"fmt"
	"os"
	"time"

	"./sdfl"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func compile(filePath string) {
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
	sdfl.Reset()
	sdfl.Generate(&program)

	// fmt.Println(sdfl.GetCode())
	f, err := os.Create("out_frag.glsl")
	check(err)
	defer f.Close()
	len, err := f.WriteString(sdfl.GetCode())
	check(err)
	fmt.Println(len, "bytes written successfully")
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("ERROR: sflc <input.sdfl> [--watch]")
		return
	}

	filePath := args[0]
	watchMode := false
	if len(args) > 1 && args[1] == "--watch" {
		watchMode = true
	}

	if watchMode {
		fw, err := sdfl.NewFileWatcher(filePath)
		check(err)

		fmt.Println("Watching", filePath, "for changes... (Ctrl+C to exit)")
		for {
			changed, err := fw.HasChanged()
			check(err)

			if changed {
				compile(filePath)
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		compile(filePath)
	}
}
