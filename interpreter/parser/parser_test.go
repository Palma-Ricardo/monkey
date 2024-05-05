package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

func TestLetStatements(tester *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}
	for _, testcase := range tests {
		lexer := lexer.New(testcase.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(tester, parser)

		if len(program.Statements) != 1 {
			tester.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		statement := program.Statements[0]
		if !testLetStatement(tester, statement, testcase.expectedIdentifier) {
			return
		}

		value := statement.(*ast.LetStatement).Value
		if !testLiteralExpression(tester, value, testcase.expectedValue) {
			return
		}
	}

}

func TestReturnStatements(tester *testing.T) {
	input := `
return 5;
return 10;
return 9944322;
`

	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	if len(program.Statements) != 3 {
		tester.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	for _, statement := range program.Statements {
		returnStatement, ok := statement.(*ast.ReturnStatement)
		if !ok {
			tester.Errorf("statement is not *ast.ReturnStatement. got=%T", statement)
		}

		if returnStatement.TokenLiteral() != "return" {
			tester.Errorf("returnStatement.TokenLiteral not 'return', got %q",
				returnStatement.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(tester *testing.T) {
	input := "foobar;"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	if len(program.Statements) != 1 {
		tester.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		tester.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	identifier, ok := statement.Expression.(*ast.Identifier)
	if !ok {
		tester.Fatalf("expression is not *ast.Identifier. got=%T", statement.Expression)
	}
	if identifier.Value != "foobar" {
		tester.Errorf("identifier.Value not %s. got=%s", "foobar", identifier.Value)
	}
	if identifier.TokenLiteral() != "foobar" {
		tester.Errorf("identifier.TokenLiteral not %s. got=%s",
			"foobar", identifier.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(tester *testing.T) {
	input := "5;"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	if len(program.Statements) != 1 {
		tester.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		tester.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := statement.Expression.(*ast.IntegerLiteral)
	if !ok {
		tester.Fatalf("expressions is not *ast.IntegerLiteral. got=%T", statement.Expression)
	}

	if literal.Value != 5 {
		tester.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		tester.Errorf("literal.TokenLiteral not %s. got=%s",
			"5", literal.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(tester *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, testcase := range prefixTests {
		lexer := lexer.New(testcase.input)
		parse := New(lexer)
		program := parse.ParseProgram()

		if len(program.Statements) != 1 {
			tester.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			tester.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		expression, ok := statement.Expression.(*ast.PrefixExpression)
		if !ok {
			tester.Fatalf("statement is not ast.PrefixExpression. got=%T", statement.Expression)
		}
		if expression.Operator != testcase.operator {
			tester.Fatalf("expression.Operator is not '%s'. got=%s",
				testcase.operator, expression.Operator)
		}

		if !testLiteralExpression(tester, expression.Right, testcase.integerValue) {
			return
		}
	}
}

func TestParsingInfixExpression(tester *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, testcase := range infixTests {
		lexer := lexer.New(testcase.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(tester, parser)

		if len(program.Statements) != 1 {
			tester.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			tester.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		if !testInfixExpression(tester, statement.Expression, testcase.leftValue, testcase.operator, testcase.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(tester *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)"},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, testcase := range tests {
		lexer := lexer.New(testcase.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(tester, parser)

		actual := program.String()
		if actual != testcase.expected {
			tester.Errorf("expected=%q, got=%q", testcase.expected, actual)
		}
	}
}

func TestIfExpression(tester *testing.T) {
	input := `if (x < y) { x }`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	if len(program.Statements) != 1 {
		tester.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		tester.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	expression, ok := statement.Expression.(*ast.IfExpression)
	if !ok {
		tester.Fatalf("statement.Expression is not ast.IfExpression. got=%T",
			statement.Expression)
	}

	if !testInfixExpression(tester, expression.Condition, "x", "<", "y") {
		return
	}

	if len(expression.Consequence.Statements) != 1 {
		tester.Errorf("consequence is not 1 statements. got=%d\n",
			len(expression.Consequence.Statements))
	}

	consequence, ok := expression.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		tester.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			expression.Consequence.Statements[0])
	}

	if !testIdentifier(tester, consequence.Expression, "x") {
		return
	}

	if expression.Alternative != nil {
		tester.Errorf("expression.Alternative was not nil. got=%+v", expression.Alternative)
	}
}

func TestFunctionLiteralParsing(tester *testing.T) {
	input := `fn(x, y) { x + y; }`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	if len(program.Statements) != 1 {
		tester.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		tester.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := statement.Expression.(*ast.FunctionLiteral)
	if !ok {
		tester.Fatalf("statement.Expression is not ast.FunctionLiteral. got=%T",
			statement.Expression)
	}

	if len(function.Parameters) != 2 {
		tester.Fatalf("function literal parameters wrong. want 2, got=%d\n",
			len(function.Parameters))
	}

	testLiteralExpression(tester, function.Parameters[0], "x")
	testLiteralExpression(tester, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		tester.Fatalf("function.Body.Statements does not have 1 statement. got=%d\n",
			len(function.Body.Statements))
	}

	bodyStatement, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		tester.Fatalf("function body statement is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}

	testInfixExpression(tester, bodyStatement.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(tester *testing.T) {
	tests := []struct {
		input              string
		expectedParameters []string
	}{
		{input: "fn() {};", expectedParameters: []string{}},
		{input: "fn(x) {};", expectedParameters: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParameters: []string{"x", "y", "z"}},
	}

	for _, testcase := range tests {
		lexer := lexer.New(testcase.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(tester, parser)

		statement := program.Statements[0].(*ast.ExpressionStatement)
		function := statement.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(testcase.expectedParameters) {
			tester.Fatalf("number of parameters wrong. want %d, got=%d\n",
				len(testcase.expectedParameters), len(function.Parameters))
		}

		for i, identifier := range testcase.expectedParameters {
			testLiteralExpression(tester, function.Parameters[i], identifier)
		}
	}
}

func TestCallExpressionParsing(tester *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	if len(program.Statements) != 1 {
		tester.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		tester.Fatalf("statement is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	expression, ok := statement.Expression.(*ast.CallExpression)
	if !ok {
		tester.Fatalf("statement.Expression is not ast.CallExpression. got=%T",
			statement.Expression)
	}

	if !testIdentifier(tester, expression.Function, "add") {
		return
	}

	if len(expression.Arguments) != 3 {
		tester.Fatalf("wrong number of arguments. got=%d", len(expression.Arguments))
	}

	testLiteralExpression(tester, expression.Arguments[0], 1)
	testInfixExpression(tester, expression.Arguments[1], 2, "*", 3)
	testInfixExpression(tester, expression.Arguments[2], 4, "+", 5)
}

func TestStringLiteralExpression(tester *testing.T) {
	input := `"hello world";`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := statement.Expression.(*ast.StringLiteral)
	if !ok {
		tester.Fatalf("expressions is not *ast.StringLiteral. got=%T", statement.Expression)
	}

	if literal.Value != "hello world" {
		tester.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestParsingArrayLiterals(tester *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := statement.Expression.(*ast.ArrayLiteral)
	if !ok {
		tester.Fatalf("expression is not ast.ArrayLiteral. got=%T", statement.Expression)
	}

	if len(array.Elements) != 3 {
		tester.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
	}

	testIntegerLiteral(tester, array.Elements[0], 1)
	testInfixExpression(tester, array.Elements[1], 2, "*", 2)
	testInfixExpression(tester, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(tester *testing.T) {
	input := "myArray[1 + 1]"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	indexExpression, ok := statement.Expression.(*ast.IndexExpression)
	if !ok {
		tester.Fatalf("expression is not *ast.IndexExpression. got=%T", statement.Expression)
	}

	if !testIdentifier(tester, indexExpression.Left, "myArray") {
		return
	}

	if !testInfixExpression(tester, indexExpression.Index, 1, "+", 1) {
		return
	}
}

func TestParsingHashLiteralsStringKeys(tester *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := statement.Expression.(*ast.HashLiteral)
	if !ok {
		tester.Fatalf("expression is not *ast.HashLiteral. got=%T", statement.Expression)
	}

	if len(hash.Pairs) != 3 {
		tester.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			tester.Errorf("key is not ast.StringLiteral. got=%T", key)
		}

		expectedValue := expected[literal.String()]

		testIntegerLiteral(tester, value, expectedValue)
	}
}

func TestParsingEmptyHashLiteral(tester *testing.T) {
	input := "{}"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()
	checkParserErrors(tester, parser)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := statement.Expression.(*ast.HashLiteral)
	if !ok {
		tester.Fatalf("expression is not ast.HashLiteral. got=%T", statement.Expression)
	}

	if len(hash.Pairs) != 0 {
		tester.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
}

func testLetStatement(tester *testing.T, statement ast.Statement, name string) bool {
	if statement.TokenLiteral() != "let" {
		tester.Errorf("statement.TokenLiteral not 'let'. got=%q", statement.TokenLiteral())
		return false
	}

	letStatement, ok := statement.(*ast.LetStatement)
	if !ok {
		tester.Errorf("statement is not *ast.LetStatement. got=%T", statement)
		return false
	}

	if letStatement.Name.Value != name {
		tester.Errorf("letStatement.Name.Value not '%s', got=%s", name, letStatement.Name.Value)
		return false
	}

	if letStatement.Name.TokenLiteral() != name {
		tester.Errorf("letStatement.Name.TokenLiteral() not '%s'. got=%s",
			name, letStatement.Name.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(tester *testing.T, il ast.Expression, value int64) bool {
	integer, ok := il.(*ast.IntegerLiteral)
	if !ok {
		tester.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integer.Value != value {
		tester.Errorf("integer.Value not %d. got=%d", value, integer.Value)
		return false
	}

	if integer.TokenLiteral() != fmt.Sprintf("%d", value) {
		tester.Errorf("integer.TokenLiteral not %d. got=%s", value, integer.TokenLiteral())
		return false
	}

	return true
}

func testInfixExpression(tester *testing.T, expression ast.Expression, left interface{},
	operator string, right interface{}) bool {
	operatorExpresion, ok := expression.(*ast.InfixExpression)
	if !ok {
		tester.Errorf("expression is not ast.InfixExpression. got=%T(%s)", expression, expression)
		return false
	}

	if !testLiteralExpression(tester, operatorExpresion.Left, left) {
		return false
	}

	if operatorExpresion.Operator != operator {
		tester.Errorf("expression.Operator is not '%s'. got=%q", operator, operatorExpresion.Operator)
		return false
	}

	if !testLiteralExpression(tester, operatorExpresion.Right, right) {
		return false
	}
	return true
}

func testIdentifier(tester *testing.T, expression ast.Expression, value string) bool {
	identifier, ok := expression.(*ast.Identifier)
	if !ok {
		tester.Errorf("expression is not *ast.Identifier. got=%T", expression)
		return false
	}

	if identifier.Value != value {
		tester.Errorf("identifier.Value not %s. got=%s", value, identifier.Value)
		return false
	}

	if identifier.TokenLiteral() != value {
		tester.Errorf("identifier.TokenLiteral not %s. got=%s", value, identifier.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(tester *testing.T, expression ast.Expression, value bool) bool {
	boolean, ok := expression.(*ast.Boolean)
	if !ok {
		tester.Errorf("expression is not *ast.Boolean. got=%T", expression)
		return false
	}

	if boolean.Value != value {
		tester.Errorf("boolean.Value is not %t. got=%t", value, boolean.Value)
		return false
	}

	if boolean.TokenLiteral() != fmt.Sprintf("%t", value) {
		tester.Errorf("boolean.TokenLiteral is not %t. got=%s", value, boolean.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(tester *testing.T, expression ast.Expression, expected interface{}) bool {
	switch value := expected.(type) {
	case int:
		return testIntegerLiteral(tester, expression, int64(value))
	case int64:
		return testIntegerLiteral(tester, expression, value)
	case string:
		return testIdentifier(tester, expression, value)
	case bool:
		return testBooleanLiteral(tester, expression, value)
	}

	tester.Errorf("type of exp not handled. got=%T", expression)
	return false
}

func checkParserErrors(tester *testing.T, parser *Parser) {
	errors := parser.Errors()
	if len(errors) == 0 {
		return
	}

	tester.Errorf("parser has %d errors", len(errors))
	for _, message := range errors {
		tester.Errorf("parser error: %q", message)
	}
	tester.FailNow()
}
