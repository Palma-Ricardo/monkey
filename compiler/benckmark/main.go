package main

import (
	"flag"
	"fmt"
	"monkey/compiler"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")

var input = `
let fibonacci = fn(x) {
    if (x == 0) {
        0
    } else {
        if (x == 1) {
            return 1;
        } else {
            fibonacci(x - 1) + fibonacci(x - 2);
        }
    }
};
fibonacci(35);
`

func main() {
	flag.Parse()

	var duration time.Duration
	var result object.Object

	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()

	if *engine == "vm" {
		compiler := compiler.New()
		error := compiler.Compile(program)
		if error != nil {
			fmt.Printf("compiler error: %s", error)
			return
		}

		machine := vm.New(compiler.Bytecode())

		start := time.Now()

		error = machine.Run()
		if error != nil {
			fmt.Printf("vm error: %s", error)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		result = evaluator.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf("engine=%s result=%s duration=%s\n", *engine, result.Inspect(), duration)
}
