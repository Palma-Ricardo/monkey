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

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
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

		stackElem := vm.LastPoppedStackElem()

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
	case bool:
		error := testBooleanObject(bool(expected), actual)
		if error != nil {
			tester.Errorf("testBooleanObject failed: %s", error)
		}
	case string:
		error := testStringObject(expected, actual)
		if error != nil {
			tester.Errorf("testStringObject failed: %s", error)
		}
	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			tester.Errorf("object is not Array: %T (%+v)", actual, array)
			return
		}

		if len(array.Elements) != len(expected) {
			tester.Errorf("wrong number of elements. want=%d, got=%d", len(expected), len(array.Elements))
			return
		}

		for i, expectedElement := range expected {
			error := testIntegerObject(int64(expectedElement), array.Elements[i])
			if error != nil {
				tester.Errorf("testIntegerObject failed: %s", error)
			}
		}
	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			tester.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			tester.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(expected), len(hash.Pairs))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				tester.Errorf("no pair for key in Pairs")
			}

			error := testIntegerObject(expectedValue, pair.Value)
			if error != nil {
				tester.Errorf("testIntegerObject failed: %s", error)
			}
		}

	case *object.Null:
		if actual != Null {
			tester.Errorf("object is not Null: %T (%+v)", actual, actual)
		}
	}
}

func TestIntegerArithmetic(tester *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(tester, tests)
}

func TestBooleanExpressions(tester *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!(if (false) { 5; })", true},
	}

	runVmTests(tester, tests)
}

func TestConditions(tester *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) { 10 }", Null},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}

	runVmTests(tester, tests)
}

func TestGlobalLetStatements(tester *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVmTests(tester, tests)
}

func TestStringExpressions(tester *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}

	runVmTests(tester, tests)
}

func TestArrayLiterals(tester *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 - 4, 5 + 6]", []int{3, -1, 11}},
	}

	runVmTests(tester, tests)
}

func TestHashLiterals(tester *testing.T) {
	tests := []vmTestCase{
		{
			"{}", map[object.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVmTests(tester, tests)
}

func TestIndexExpressions(tester *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Null},
		{"[1, 2, 3][99]", Null},
		{"[1][-1]", Null},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Null},
		{"{}[0]", Null},
	}

	runVmTests(tester, tests)
}

func TestCallingFunctionWithoutArguments(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let fivePlusTen = fn() { 5 + 10; }; fivePlusTen();",
			expected: 15,
		},
		{
			input:    "let one = fn() { 1; }; let two = fn() { 2; }; one() + two()",
			expected: 3,
		},
		{
			input:    "let a = fn() { 1 }; let b = fn() { a() + 1 }; let c = fn() { b() + 1}; c();",
			expected: 3,
		},
	}

	runVmTests(tester, tests)
}

func TestFunctionsWithReturnStatement(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let earlyExit = fn() {return 99; 100; }; earlyExit();",
			expected: 99,
		},
		{
			input:    "let earlyExit = fn() {return 99; return 100; }; earlyExit();",
			expected: 99,
		},
	}

	runVmTests(tester, tests)
}

func TestFunctionsWithoutReturnValue(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let noReturn = fn() { }; noReturn();",
			expected: Null,
		},
		{
			input:    "let noReturn = fn() { }; let noReturnTwo = fn() { noReturn(); }; noReturn(); noReturnTwo();",
			expected: Null,
		},
	}

	runVmTests(tester, tests)
}
