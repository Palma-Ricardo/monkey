package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type Parser struct {
	lexer  *lexer.Lexer
	errors []string

	currentToken token.Token
	peekToken    token.Token

	prefixParseFunctions map[token.TokenType]prefixParseFunction
	infixParseFunctions  map[token.TokenType]infixParseFunction
}

func New(lexer *lexer.Lexer) *Parser {
	parser := &Parser{
		lexer:  lexer,
		errors: []string{},
	}

	parser.prefixParseFunctions = make(map[token.TokenType]prefixParseFunction)
	parser.registerPrefix(token.IDENT, parser.parseIdentifier)
	parser.registerPrefix(token.INT, parser.parseIntegerLiteral)
	parser.registerPrefix(token.BANG, parser.parsePrefixExpression)
	parser.registerPrefix(token.MINUS, parser.parsePrefixExpression)
	parser.registerPrefix(token.TRUE, parser.parseBoolean)
	parser.registerPrefix(token.FALSE, parser.parseBoolean)
	parser.registerPrefix(token.LPAREN, parser.parseGroupedExpression)
	parser.registerPrefix(token.IF, parser.parseIfExpression)

	parser.infixParseFunctions = make(map[token.TokenType]infixParseFunction)
	parser.registerInfix(token.PLUS, parser.parseInfixExpression)
	parser.registerInfix(token.MINUS, parser.parseInfixExpression)
	parser.registerInfix(token.SLASH, parser.parseInfixExpression)
	parser.registerInfix(token.STAR, parser.parseInfixExpression)
	parser.registerInfix(token.EQUAL, parser.parseInfixExpression)
	parser.registerInfix(token.NOTEQUAL, parser.parseInfixExpression)
	parser.registerInfix(token.LESS, parser.parseInfixExpression)
	parser.registerInfix(token.GREATER, parser.parseInfixExpression)

	parser.nextToken()
	parser.nextToken()
	return parser
}

func (parser *Parser) Errors() []string {
	return parser.errors
}

func (parser *Parser) peekError(t token.TokenType) {
	message := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, parser.peekToken.Type)

	parser.errors = append(parser.errors, message)
}

func (parser *Parser) nextToken() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) registerPrefix(tokenType token.TokenType, function prefixParseFunction) {
	parser.prefixParseFunctions[tokenType] = function
}

func (parser *Parser) registerInfix(tokenType token.TokenType, function infixParseFunction) {
	parser.infixParseFunctions[tokenType] = function
}

func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for parser.currentToken.Type != token.EOF {
		statement := parser.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		parser.nextToken()
	}

	return program
}

func (parser *Parser) parseStatement() ast.Statement {
	switch parser.currentToken.Type {
	case token.LET:
		return parser.parseLetStatement()
	case token.RETURN:
		return parser.parseReturnStatement()
	default:
		return parser.parseExpressionStatement()
	}
}

func (parser *Parser) parseLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: parser.currentToken}

	if !parser.expectPeek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}

	if !parser.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO
	for !parser.currentTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: parser.currentToken}

	parser.nextToken()

	// TODO
	for !parser.currentTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: parser.currentToken}
	statement.Expression = parser.parseExpression(LOWEST)

	if parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseExpression(precedence int) ast.Expression {
	prefix := parser.prefixParseFunctions[parser.currentToken.Type]
	if prefix == nil {
		parser.noPrefixParseFunctionError(parser.currentToken.Type)
		return nil
	}
	leftExpression := prefix()

	for !parser.peekTokenIs(token.SEMICOLON) && precedence < parser.peekPrecedence() {
		infix := parser.infixParseFunctions[parser.peekToken.Type]
		if infix == nil {
			return leftExpression
		}

		parser.nextToken()

		leftExpression = infix(leftExpression)
	}

	return leftExpression
}

func (parser *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: parser.currentToken}

	value, err := strconv.ParseInt(parser.currentToken.Literal, 0, 64)
	if err != nil {
		message := fmt.Sprintf("could not parse %q as integer", parser.currentToken.Literal)
		parser.errors = append(parser.errors, message)
		return nil
	}

	literal.Value = value
	return literal
}

func (parser *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: parser.currentToken, Value: parser.currentToken.Literal}
}

func (parser *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    parser.currentToken,
		Operator: parser.currentToken.Literal,
	}

	parser.nextToken()

	expression.Right = parser.parseExpression(PREFIX)

	return expression
}

func (parser *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    parser.currentToken,
		Operator: parser.currentToken.Literal,
		Left:     left,
	}

	precedence := parser.currentPrecedence()
	parser.nextToken()
	expression.Right = parser.parseExpression(precedence)

	return expression
}

func (parser *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: parser.currentToken, Value: parser.currentTokenIs(token.TRUE)}
}

func (parser *Parser) parseGroupedExpression() ast.Expression {
	parser.nextToken()

	expression := parser.parseExpression(LOWEST)

	if !parser.expectPeek(token.RPAREN) {
		return nil
	}

	return expression
}

func (parser *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: parser.currentToken}

	if !parser.expectPeek(token.LPAREN) {
		return nil
	}

	parser.nextToken()
	expression.Condition = parser.parseExpression(LOWEST)

	if !parser.expectPeek(token.RPAREN) {
		return nil
	}

	if !parser.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = parser.parseBlockStatement()

    if parser.peekTokenIs(token.ELSE) {
        parser.nextToken()

        if !parser.expectPeek(token.LBRACE) {
            return nil
        }

        expression.Alternative = parser.parseBlockStatement()
    }

	return expression
}

func (parser *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: parser.currentToken}
	block.Statements = []ast.Statement{}

	parser.nextToken()

	for !parser.currentTokenIs(token.RBRACE) && !parser.currentTokenIs(token.EOF) {
		statement := parser.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
		parser.nextToken()
	}

	return block
}

func (parser *Parser) currentTokenIs(t token.TokenType) bool {
	return parser.currentToken.Type == t
}

func (parser *Parser) peekTokenIs(t token.TokenType) bool {
	return parser.peekToken.Type == t
}

func (parser *Parser) expectPeek(t token.TokenType) bool {
	if parser.peekTokenIs(t) {
		parser.nextToken()
		return true
	} else {
		parser.peekError(t)
		return false
	}
}

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +, -
	PRODUCT     // *, /
	PREFIX      // -value or !value
	CALL        // function(value)
)

var precedences = map[token.TokenType]int{
	token.EQUAL:    EQUALS,
	token.NOTEQUAL: EQUALS,
	token.LESS:     LESSGREATER,
	token.GREATER:  LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.STAR:     PRODUCT,
	token.SLASH:    PRODUCT,
}

type (
	prefixParseFunction func() ast.Expression
	infixParseFunction  func(ast.Expression) ast.Expression
)

func (parser *Parser) peekPrecedence() int {
	if precedence, ok := precedences[parser.peekToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (parser *Parser) currentPrecedence() int {
	if precedence, ok := precedences[parser.currentToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (parser *Parser) noPrefixParseFunctionError(t token.TokenType) {
	message := fmt.Sprintf("no prefix parse function for %s found", t)
	parser.errors = append(parser.errors, message)
}
