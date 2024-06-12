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
			fmt.Printf("Input: %s", testcase.input)
			tester.Fatalf("testInstructions failed: %s", error)
		}

		error = testConstants(testcase.expectedConstants, bytecode.Constants)
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

func testConstants(expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d",
			len(actual), len(expected))
	}

	for index, constant := range expected {
		switch constant := constant.(type) {
		case int:
			error := testIntegerObject(int64(constant), actual[index])
			if error != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s",
					index, error)
			}
		case string:
			error := testStringObject(constant, actual[index])
			if error != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s",
					index, error)
			}
		case []code.Instructions:
			fn, ok := actual[index].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", index, actual[index])
			}

			error := testInstructions(constant, fn.Instructions)
			if error != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", index, error)
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

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
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
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 * 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "2 / 1",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestBooleanExpressions(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestConditionals(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if (true) { 10 }; 3333;",
			expectedConstants: []interface{}{10, 3333},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpJumpNotTrue, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 11),
				code.Make(code.OpNull),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "if (true) { 10 } else { 20 }; 3333;",
			expectedConstants: []interface{}{10, 20, 3333},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpJumpNotTrue, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 13),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestGlobalLetStatements(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "let one = 1; let two = 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input:             "let one = 1; one;",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "let one = 1; let two = one; two;",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestStringExpressions(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"monkey"`,
			expectedConstants: []interface{}{"monkey"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"mon" + "key"`,
			expectedConstants: []interface{}{"mon", "key"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestArrayLiterals(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestHashLiterals(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpHash, 6),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestIndexExpressions(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []interface{}{1, 2, 2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestFunctions(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { return 5 + 10 }",
			expectedConstants: []interface{}{
				5, 10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { 5 + 10 }",
			expectedConstants: []interface{}{
				5, 10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { 1; 2 }",
			expectedConstants: []interface{}{
				1, 2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestCompilerScopes(tester *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		tester.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}
	globalSymbolTable := compiler.symbolTable

	compiler.emit(code.OpMul)

	compiler.enterScope()
	if compiler.scopeIndex != 1 {
		tester.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 1)
	}

	compiler.emit(code.OpSub)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		tester.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last := compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpSub {
		tester.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpSub)
	}

	if compiler.symbolTable.Outer != globalSymbolTable {
		tester.Errorf("compiler did not enclose symbolTable")
	}

	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		tester.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}

	if compiler.symbolTable != globalSymbolTable {
		tester.Errorf("compiler did not restore global symbol table")
	}
	if compiler.symbolTable.Outer != nil {
		tester.Errorf("compiler modified global symbol table incorrectly")
	}

	compiler.emit(code.OpAdd)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		tester.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last = compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		tester.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpAdd)
	}

	previous := compiler.scopes[compiler.scopeIndex].previousInstruction
	if previous.Opcode != code.OpMul {
		tester.Errorf("previousInstruction.Opcode wrong. got=%d, want=%d", previous.Opcode, code.OpMul)
	}
}

func TestFunctionsWithoutReturnValue(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { }",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestFunctionCalls(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { 24 }();",
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let noArg = fn() { 24 }; noArg();",
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let oneArg = fn(a) { a }; oneArg(24);`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let manyArg = fn(a, b, c) { a; b; c }; manyArg(24, 25, 26);",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpReturnValue),
				},
				24, 25, 26,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestLetStatementScopes(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input: "let num = 55; fn() { num }",
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { let num = 55; num }",
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { let a = 55; let b = 77; a + b }",
			expectedConstants: []interface{}{
				55, 77,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestBuiltins(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
            len([]);
            push([], 1);
            `,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpArray, 0),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
				code.Make(code.OpGetBuiltin, 5),
				code.Make(code.OpArray, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { len([]) }`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpArray, 0),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}

func TestClosures(tester *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn(a) { fn(b) { a + b } }",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn(a) { fn(b) { fn(c) { a + b + c } } }",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 1, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            let global = 55;
            fn() {
            let a = 66;
            fn() {
                let b = 77;
                    fn() {
                        let c = 88;
                        global + a + b + c;
                    }
                }
            }
            `,
			expectedConstants: []interface{}{
				55, 66, 77, 88,
				[]code.Instructions{
					code.Make(code.OpConstant, 3),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 2),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 4, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 5, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 6, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(tester, tests)
}
