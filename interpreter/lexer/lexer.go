package lexer

import "monkey/token"

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

func (lexer *Lexer) readChar() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.ch = 0
	} else {
		lexer.ch = lexer.input[lexer.readPosition]
	}
	lexer.position = lexer.readPosition
	lexer.readPosition += 1
}

func (lexer *Lexer) NextToken() token.Token {
	var tok token.Token

	switch lexer.ch {
    case ';':
        tok = newToken(token.SEMICOLON, lexer.ch)
    case '(':
        tok = newToken(token.LPAREN, lexer.ch)
    case ')':
        tok = newToken(token.RPAREN, lexer.ch)
    case '{':
        tok = newToken(token.LBRACE, lexer.ch)
    case '}':
        tok = newToken(token.RBRACE, lexer.ch)
    case ',':
        tok = newToken(token.COMMA, lexer.ch)
	case '=':
		tok = newToken(token.ASSIGN, lexer.ch)
    case '+':
        tok = newToken(token.PLUS, lexer.ch)
    case 0:
        tok.Literal = ""
        tok.Type = token.EOF
	}

    lexer.readChar()
    return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
