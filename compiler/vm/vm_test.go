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

		for i, constant := range compiler.Bytecode().Constants {
			fmt.Printf("CONSTANT %d %p (%T):\n", i, constant, constant)

			switch constant := constant.(type) {
			case *object.CompiledFunction:
				fmt.Printf(" Instructions:\n%s", constant.Instructions)
			case *object.Integer:
				fmt.Printf(" Value: %d\n", constant.Value)
			}

			fmt.Printf("\n")
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
	case *object.Error:
		errorObject, ok := actual.(*object.Error)
		if !ok {
			tester.Errorf("object is not Error: %T (%+v)", actual, actual)
		}

		if errorObject.Message != expected.Message {
			tester.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errorObject.Message)
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

func TestFirstClassFunctions(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let returnsOne = fn() { 1; }; let returnsOneReturner = fn() { returnsOne; }; returnsOneReturner()();",
			expected: 1,
		},
		{
			input:    "let returnsOneReturner = fn() { let returnsOne = fn() { 1; };  returnsOne;}; returnsOneReturner()();",
			expected: 1,
		},
	}

	runVmTests(tester, tests)
}

func TestCallingFunctionsWithBindings(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let one = fn() { let one = 1; one }; one();",
			expected: 1,
		},
		{
			input: `let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
            let threeAndFour = fn() { let three = 3; let four = 4; three + 4; };
            oneAndTwo() + threeAndFour();`,
			expected: 10,
		},
		{
			input: `let firstFoobar = fn() { let foobar = 50; foobar; };
            let secondFoobar = fn() { let foobar = 100; foobar; };
            firstFoobar() + secondFoobar();`,
			expected: 150,
		},
		{
			input: `let globalSeed = 50; let minusOne = fn() { let num = 1; globalSeed - num;}
            let minusTwo = fn() { let num = 2; globalSeed - num; }
            minusOne() + minusTwo();`,
			expected: 97,
		},
	}

	runVmTests(tester, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let identity = fn(a) { a; }; identity(4);",
			expected: 4,
		},
		{
			input:    "let sum = fn(a, b) { a + b; }; sum(1, 2);",
			expected: 3,
		},
		{
			input:    "let sum = fn(a, b) { let c = a + b; c;}; sum(1, 2);",
			expected: 3,
		},
		{
			input:    "let sum = fn(a, b) { let c = a + b; c; }; sum(1, 2) + sum(3, 4);",
			expected: 10,
		},
		{
			input:    "let sum = fn(a, b) { let c = a + b; c; }; let outer = fn() { sum(1, 2) + sum(3, 4); }; outer();",
			expected: 10,
		},
		{
			input: `let globalNum = 10; let sum = fn(a, b) { let c = a + b; c + globalNum; };
            let outer = fn() { sum(1, 2) + sum(3, 4) + globalNum; }; outer() + globalNum;`,
			expected: 50,
		},
	}

	runVmTests(tester, tests)
}

func TestCallingFunctionsWithWrongArguments(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "fn () { 1; }(1);",
			expected: "wrong number of arguments: want=0, got=1",
		},
		{
			input:    "fn(a) { a; }();",
			expected: "wrong number of arguments: want=1, got=0",
		},
		{
			input:    "fn(a, b) { a + b; }(1);",
			expected: "wrong number of arguments: want=2, got=1",
		},
	}

	for _, testcase := range tests {
		program := parse(testcase.input)

		comp := compiler.New()
		error := comp.Compile(program)
		if error != nil {
			tester.Fatalf("compiler error: %s", error)
		}

		vm := New(comp.Bytecode())
		error = vm.Run()
		if error == nil {
			tester.Fatalf("expected VM error but resulted in none.")
		}

		if error.Error() != testcase.expected {
			tester.Fatalf("wrong VM error: want=%q, got=%q", testcase.expected, error)
		}
	}
}

func TestBuiltinFunctions(tester *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			&object.Error{
				Message: "argument to `len` not supported, got INTEGER",
			},
		},
		{`len("one", "two")`,
			&object.Error{
				Message: "wrong number of arguments. got=2, want=1",
			},
		},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`puts("hello", "world!")`, Null},
		{`first([1, 2, 3])`, 1},
		{`first([])`, Null},
		{`first(1)`,
			&object.Error{
				Message: "argument to `first` must be ARRAY, got INTEGER",
			},
		},
		{`last([1, 2, 3])`, 3},
		{`last([])`, Null},
		{`last(1)`,
			&object.Error{
				Message: "argument to `last` must be ARRAY, got INTEGER",
			},
		},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, Null},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`,
			&object.Error{
				Message: "argument to `push` must be ARRAY, got INTEGER",
			},
		},
	}

	runVmTests(tester, tests)
}

func TestClosures(tester *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let newClosure = fn(a) { fn() { a; }; }; let closure = newClosure(99); closure();",
			expected: 99,
		},
		{
			input: `
            let newAdder = fn(a, b) {
                fn(c) { a + b + c };
            };
            let adder = newAdder(1, 2);
            adder(8);
            `,
			expected: 11,
		},
		{
			input: `
            let newAdder = fn(a, b) {
                let c = a + b;
                fn(d) { c + d };
            };
            let adder = newAdder(1, 2);
            adder(8);
            `,
			expected: 11,
		},
		{
			input: `
            let newAdderOuter = fn(a, b) {
                let c = a + b;
                fn(d) {
                    let e = d + c;
                    fn(f) { e + f; };
                };
            };
            let newAdderInner = newAdderOuter(1, 2)
            let adder = newAdderInner(3);
            adder(8);
            `,
			expected: 14,
		},
		{
			input: `
            let a = 1;
            let newAdderOuter = fn(b) {
                fn(c) {
                    fn(d) { a + b + c + d };
                };
            };
            let newAdderInner = newAdderOuter(2)
            let adder = newAdderInner(3);
            adder(8);
            `,
			expected: 14,
		},
		{
			input: `
            let newClosure = fn(a, b) {
                let one = fn() { a; };
                let two = fn() { b; };
                fn() { one() + two(); };
            };
            let closure = newClosure(9, 90);
            closure();
            `,
			expected: 99,
		},
	}

	runVmTests(tester, tests)
}

func TestRecursiveFunctions(tester *testing.T) {
	tests := []vmTestCase{
		{
			input: `
            let countDown = fn(x) {
                if (x == 0) {
                    return 0;
                } else {
                    countDown(x - 1);
                }
            };
            countDown(1);
            `,
			expected: 0,
		},
		{
			input: `
            let countDown = fn(x) {
                if (x == 0) {
                    return 0;
                } else {
                    countDown(x - 1);
                }
            };
            let wrapper = fn() {
                countDown(1);
            };
            wrapper();
            `,
			expected: 0,
		},
		{
			input: `
            let wrapper = fn() {
                let countDown = fn(x) {
                    if (x == 0) {
                        return 0;
                    } else {
                        countDown(x - 1);
                    }
                }
                countDown(1);
            }
            wrapper();
            `,
			expected: 0,
		},
	}

	runVmTests(tester, tests)
}

func TestRecursiveFibonacci(tester *testing.T) {
	tests := []vmTestCase{
		{
			input: `
            let fibonacci = fn(x) {
                if (x == 0)  {
                    return 0;
                } else {
                    if (x == 1) {
                        return 1;
                    } else {
                        return fibonacci(x - 1) + fibonacci(x - 2);
                    }
                }
            }
            fibonacci(15);
            `,
			expected: 610,
		},
	}

	runVmTests(tester, tests)
}
