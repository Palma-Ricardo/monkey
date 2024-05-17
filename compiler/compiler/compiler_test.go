package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func runCompilerTests(tester *testing.T, tests []compilerTestCase) {
	tester.Helper()

	for _, testcase := range tests {
		program := parse(testcase.input)

		compiler := New()
		error := compiler.Compile(program)
		if error != nil {
			tester.Fatalf("compiler error: %s", error)
		}

		bytecode := compiler.Bytecode()

		error = testInstructions(testcase.expectedInstructions, bytecode.Instructions)
		if error != nil {
			tester.Fatalf("testInstructions failed: %s", error)
		}

		error = testConstants(tester, testcase.expectedConstants, bytecode.Constants)
		if error != nil {
			tester.Fatalf("testConstans failed: %s", error)
		}
	}
}

func parse(input string) *ast.Program {
	lexer := lexer.New(input)
	parser := parser.New(lexer)
	return parser.ParseProgram()
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatenated := concatenateInstructions(expected)

	if len(actual) != len(concatenated) {
		return fmt.Errorf("wrong instruction length.\nwant=%q\ngot=%q",
			concatenated, actual)
	}

	for index, instruction := range concatenated {
		if actual[index] != instruction {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot=%q",
				index, concatenated, actual)
		}
	}

	return nil
}

func concatenateInstructions(source []code.Instructions) code.Instructions {
	output := code.Instructions{}

	for _, instruction := range source {
		output = append(output, instruction...)
	}

	return output
}

func testConstants(tester *testing.T, expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d",
			len(actual), len(expected))
	}

	for index, constant := range expected {
		switch constant := constant.(type) {
		case int:
			error := testIntegerObject(int64(constant), actual[index])
			if error != nil {
				return fmt.Errorf("constant %d -  testIntegerObject failed: %s ",
					index, error)
			}
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object ahs wrong value. got=%d, want=%d",
			result.Value, expected)
	}

	return nil
}

func TestIntegerArithmetic(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1+2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
        {
            input: "1; 2",
            expectedConstants: []interface{}{1, 2},
            expectedInstructions: []code.Instructions{
                code.Make(code.OpConstant, 0),
                code.Make(code.OpPop),
                code.Make(code.OpConstant, 1),
                code.Make(code.OpPop),
            },
        },
	}

	runCompilerTests(tester, tests)
}
