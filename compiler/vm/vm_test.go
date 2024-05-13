package vm

import (
	"fmt"
	"monkey/ast"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func parse(input string) *ast.Program {
	lexer := lexer.New(input)
	parser := parser.New(lexer)

	return parser.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
	}

	return nil
}

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(tester *testing.T, tests []vmTestCase) {
	tester.Helper()

	for _, testcase := range tests {
		program := parse(testcase.input)

		compiler := compiler.New()
		err := compiler.Compile(program)
		if err != nil {
			tester.Fatalf("compiler error: %s", err)
		}

		vm := New(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			tester.Fatalf("vm error: %s", err)
		}

		stackElem := vm.StackTop()

		testExpectedObject(tester, testcase.expected, stackElem)
	}
}

func testExpectedObject(tester *testing.T, expected interface{}, actual object.Object) {
	tester.Helper()

	switch expected := expected.(type) {
	case int:
		error := testIntegerObject(int64(expected), actual)
		if error != nil {
			tester.Errorf("testIntegerObject failed: %s", error)
		}
	}
}

func TestIntegerArithmetic(tester *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 2}, // TODO: FIXME
	}

	runVmTests(tester, tests)
}
