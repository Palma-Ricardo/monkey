package lexer

import (
	"monkey/token"
	"testing"
)

func TestNextToken(tester *testing.T) {
	input := `=+(){},;`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	lexer := New(input)

	for i, testcase := range tests {
		token := lexer.NextToken()

		if token.Type != testcase.expectedType {
			tester.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, testcase.expectedType, token.Type)
		}

		if token.Literal != testcase.expectedLiteral {
			tester.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, testcase.expectedLiteral, token.Literal)
		}
	}
}
