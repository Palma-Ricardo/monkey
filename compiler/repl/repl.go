package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
)

const PROMPT = ">> "
const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		parser := parser.New(lexer)

		program := parser.ParseProgram()
		if len(parser.Errors()) != 0 {
			printParserErrors(out, parser.Errors())
			continue
		}

		compiler := compiler.NewWithState(symbolTable, constants)
		error := compiler.Compile(program)
		if error != nil {
			fmt.Fprintf(out, "Whoops! Compilation failed:\n %s\n", error)
			continue
		}

		code := compiler.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalsStore(code, globals)
		error = machine.Run()
		if error != nil {
			fmt.Fprintf(out, "Whoops! Executing bytecode failed:\n %s\n", error)
			continue
		}

		lastPoppedItem := machine.LastPoppedStackElem()
		io.WriteString(out, lastPoppedItem.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, "  parser errors:\n")
	for _, message := range errors {
		io.WriteString(out, "\t"+message+"\n")
	}
}
