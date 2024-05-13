package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func TestEvalIntegerExpression(tester *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)
		testIntegerObject(tester, evaluated, testcase.expected)
	}
}

func TestEvalBooleanExpression(tester *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
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
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)
		testBooleanObject(tester, evaluated, testcase.expected)
	}
}

func TestBangOperator(tester *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)
		testBooleanObject(tester, evaluated, testcase.expected)
	}
}

func TestIfElseExpression(tester *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)
		integer, ok := testcase.expected.(int)
		if ok {
			testIntegerObject(tester, evaluated, int64(integer))
		} else {
			testNullObject(tester, evaluated)
		}
	}
}

func TestReturnStatements(tester *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`
            if (10 > 1) {
                if (10 > 1) {
                    return 10;
                }
                return 1;
            }
            `,
			10,
		},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)
		testIntegerObject(tester, evaluated, testcase.expected)
	}
}

func TestErrorHandling(tester *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
            if (10 > 1) {
                if (10 > 1) {
                    return true + false;
                }
                return 1;
            }
            `,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			`{"name": "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)

		errorObject, ok := evaluated.(*object.Error)
		if !ok {
			tester.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errorObject.Message != testcase.expectedMessage {
			tester.Errorf("wrong error message. expected=%q, got=%q",
				testcase.expectedMessage, errorObject.Message)
		}
	}
}

func TestLetStatements(tester *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, testcase := range tests {
		testIntegerObject(tester, testEval(testcase.input), testcase.expected)
	}
}

func TestFunctionObject(tester *testing.T) {
	input := "fn(x) {x + 2;};"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		tester.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		tester.Fatalf("function has wrong number of parameters. Parameters=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		tester.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		tester.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionCalling(tester *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, testcase := range tests {
		testIntegerObject(tester, testEval(testcase.input), testcase.expected)
	}
}

func TestStringLiteral(tester *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		tester.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		tester.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(tester *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		tester.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		tester.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestBuiltinFunctions(tester *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)

		switch expected := testcase.expected.(type) {
		case int:
			testIntegerObject(tester, evaluated, int64(expected))
		case string:
			errorObject, ok := evaluated.(*object.Error)
			if !ok {
				tester.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if errorObject.Message != expected {
				tester.Errorf("wrong error message. expected=%q, got=%q",
					expected, errorObject.Message)
			}
		}
	}
}

func TestArrayLiterals(tester *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		tester.Fatalf("object is not an Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		tester.Fatalf("array has wrong number of elements. got=%d", len(result.Elements))
	}

	testIntegerObject(tester, result.Elements[0], 1)
	testIntegerObject(tester, result.Elements[1], 4)
	testIntegerObject(tester, result.Elements[2], 6)
}

func TestArrayIndexExpressions(tester *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)
		integer, ok := testcase.expected.(int)
		if ok {
			testIntegerObject(tester, evaluated, int64(integer))
		} else {
			testNullObject(tester, evaluated)
		}
	}
}

func TestHashLiterals(tester *testing.T) {
	input := `let two = "two";
    {
        "one": 10 - 9,
        two: 1 + 1,
        "thr" + "ee": 6 / 2,
        4: 4,
        true: 5,
        false: 6
    }`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		tester.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		tester.Fatalf("Hash has wrong number of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			tester.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(tester, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(tester *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, testcase := range tests {
		evaluated := testEval(testcase.input)
		integer, ok := testcase.expected.(int)
		if ok {
			testIntegerObject(tester, evaluated, int64(integer))
		} else {
			testNullObject(tester, evaluated)
		}
	}
}

func testEval(input string) object.Object {
	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()
	envirnonment := object.NewEnvironment()

	return Eval(program, envirnonment)
}

func testIntegerObject(tester *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		tester.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		tester.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(tester *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		tester.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		tester.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}

	return true
}

func testNullObject(tester *testing.T, obj object.Object) bool {
	if obj != NULL {
		tester.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}

	return true
}
