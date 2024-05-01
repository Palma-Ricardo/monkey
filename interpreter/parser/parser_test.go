package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

func TestLetStatements(tester *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
`
	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(tester, parser)
	if program == nil {
		tester.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		tester.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, testcase := range tests {
		statement := program.Statements[i]
		if !testLetStatement(tester, statement, testcase.expectedIdentifier) {
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
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
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
		if !testIntegerLiteral(tester, expression.Right, testcase.integerValue) {
			return
		}
	}
}

func TestParsingInfixExpression(tester *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
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

		testInfixExpression(tester, statement.Expression, testcase.leftValue, testcase.operator, testcase.rightValue)
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

func testLiteralExpression(tester *testing.T, expression ast.Expression, expected interface{}) bool {
	switch value := expected.(type) {
	case int:
		return testIntegerLiteral(tester, expression, int64(value))
	case int64:
		return testIntegerLiteral(tester, expression, value)
	case string:
		return testIdentifier(tester, expression, value)
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
