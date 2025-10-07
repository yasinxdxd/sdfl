package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"./sdfl"
)

type Args struct {
	args []string
	pos  int
}

func NewArgs(args []string) *Args {
	return &Args{args: args, pos: 0}
}

func (a *Args) HasNext() bool {
	return a.pos < len(a.args)
}

func (a *Args) GetNext() string {
	if !a.HasNext() {
		return ""
	}
	arg := a.args[a.pos]
	a.pos++
	return arg
}

func (a *Args) IsFlag(arg string) bool {
	return strings.HasPrefix(arg, "-")
}

func (a *Args) ParseFlag(flag string) (string, string) {
	if strings.Contains(flag, "=") {
		parts := strings.SplitN(flag, "=", 2)
		return parts[0], parts[1]
	}
	return flag, ""
}

type Config struct {
	FilePath  string
	WatchMode bool
	FromSeq   bool
	Interval  int
	ShowHelp  bool
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func compileFromSeq(filePath string) {
	sdfl.ParseSeq(filePath)
	program := sdfl.Seq2AST()
	sdfl.PrintAST(program)

	sdfl.Reset()
	sdfl.Generate(&program)

	f, err := os.Create("out_frag.glsl")
	check(err)
	defer f.Close()
	len, err := f.WriteString(sdfl.GetFragmentCode())
	check(err)
	fmt.Println(len, "bytes written successfully")

	genComputeShader()
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

	if parser.IsThereError() {
		fmt.Printf("ERROR: Syntax error!\n")
		return
	}

	sdfl.PrintAST(program)
	// convert to sequence
	sequence := sdfl.AST2Seq(program)
	// fmt.Println(sequence)
	err = sdfl.WriteSequenceToFile(program, "ast_sequence.txt", sequence)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}

	sdfl.Reset()
	sdfl.Generate(&program)
	// fmt.Println(sdfl.GetCode())
	f, err := os.Create("out_frag.glsl")
	check(err)
	defer f.Close()
	len, err := f.WriteString(sdfl.GetFragmentCode())
	check(err)
	fmt.Println(len, "bytes written successfully")

	genComputeShader()
}

func genComputeShader() {
	f, err := os.Create("out_compute.glsl")
	check(err)
	defer f.Close()
	len, err := f.WriteString(sdfl.GetComputeCode())
	check(err)
	fmt.Println(len, "bytes written successfully")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	args := NewArgs(os.Args[1:])
	config := &Config{
		Interval: 1000, // default 1 second
	}

	// Parse arguments
	for args.HasNext() {
		arg := args.GetNext()

		if args.IsFlag(arg) {
			flag, value := args.ParseFlag(arg)

			switch flag {
			case "--watch", "-w":
				config.WatchMode = true
			case "--seq", "-s":
				config.FromSeq = true
			case "--interval", "-i":
				if value == "" && args.HasNext() {
					value = args.GetNext()
				}
				if value == "" {
					fmt.Fprintf(os.Stderr, "Error: %s requires a value\n", flag)
					os.Exit(1)
				}
				interval, err := strconv.Atoi(value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid interval value: %s\n", value)
					os.Exit(1)
				}
				config.Interval = interval
			case "--help", "-h":
				config.ShowHelp = true
			default:
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", flag)
				os.Exit(1)
			}
		} else {
			// Non-flag argument should be the file path
			if config.FilePath == "" {
				config.FilePath = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", arg)
				os.Exit(1)
			}
		}
	}

	if config.ShowHelp {
		printUsage()
		return
	}

	if config.FilePath == "" {
		fmt.Fprintf(os.Stderr, "Error: no input file specified\n")
		printUsage()
		os.Exit(1)
	}

	// Execute based on configuration
	if config.WatchMode {
		fw, err := sdfl.NewFileWatcher(config.FilePath)
		check(err)

		fmt.Println("Watching", config.FilePath, "for changes... (Ctrl+C to exit)")
		fmt.Printf("Check interval: %dms\n", config.Interval)
		if config.FromSeq {
			fmt.Println("Mode: compile from sequence")
		} else {
			fmt.Println("Mode: normal compile")
		}

		for {
			changed, err := fw.HasChanged()
			check(err)

			if changed {
				fmt.Println("File changed, recompiling...")
				if config.FromSeq {
					compileFromSeq(config.FilePath)
				} else {
					compile(config.FilePath)
				}
			}
			time.Sleep(time.Duration(config.Interval) * time.Millisecond)
		}
	} else {
		// Single compilation
		if config.FromSeq {
			compileFromSeq(config.FilePath)
		} else {
			compile(config.FilePath)
		}
	}
}

func printUsage() {
	fmt.Printf(`Usage: sdflc [flags] <input.sdfl>

Flags:
  --seq, -s              Compile from sequence file
  --watch, -w            Watch mode - recompile on file changes
  --interval, -i <ms>    Watch interval in milliseconds (default: 1000)
  --help, -h             Show this help

Examples:
  sdflc input.sdfl                    # Normal compile
  sdflc --seq sequence.txt            # Compile from sequence
  sdflc --watch input.sdfl            # Watch and compile
  sdflc --watch --seq --interval=500  sequence.txt  # Watch sequence with 500ms interval
  sdflc -w -s -i 2000 input.sdfl      # Short flags
`)
}
