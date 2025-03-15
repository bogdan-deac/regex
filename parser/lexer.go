package parser

import "errors"

type TokenType = int

const (
	Char TokenType = iota
	LParen
	RParen
	Star
	Maybe
	Or
	End
	Error
)

type Token struct {
	Type  TokenType
	Value rune
}
type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
	}
}

func (l *Lexer) NextToken() (Token, error) {
	if l.pos == len(l.input) {
		return Token{Type: End}, nil
	}
	c := l.input[l.pos]
	l.pos++
	switch c {
	case '(':
		return mkOpToken(LParen), nil
	case ')':
		return mkOpToken(RParen), nil
	case '*':
		return mkOpToken(Star), nil
	case '|':
		return mkOpToken(Or), nil
	case '\\':
		if l.pos >= len(l.input) {
			return Token{Type: Error}, errors.New("Found unmatched escape at the end of the sequence")
		}
		c = l.input[l.pos]
		l.pos++
		fallthrough
	default:
		return Token{Type: Char, Value: rune(c)}, nil
	}
}

func (l *Lexer) Peek() (Token, error) {
	if l.pos == len(l.input) {
		return Token{Type: End}, nil
	}
	c := l.input[l.pos]

	switch c {
	case '(':
		return mkOpToken(LParen), nil
	case ')':
		return mkOpToken(RParen), nil
	case '*':
		return mkOpToken(Star), nil
	case '?':
		return mkOpToken(Maybe), nil
	case '|':
		return mkOpToken(Or), nil
	case '\\':
		if l.pos >= len(l.input) {
			return Token{Type: Error}, errors.New("Found unmatched escape at the end of the sequence")
		}
		c = l.input[l.pos]
		l.pos++
		fallthrough
	default:
		return Token{Type: Char, Value: rune(c)}, nil
	}
}

func mkOpToken(t TokenType) Token {
	switch t {
	case LParen:
		return Token{Type: t, Value: '('}
	case RParen:
		return Token{Type: t, Value: ')'}
	case Star:
		return Token{Type: t, Value: '*'}
	case Maybe:
		return Token{Type: t, Value: '?'}
	case Or:
		return Token{Type: t, Value: '|'}
	}
	return Token{Type: Error}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	newToken, err := l.NextToken()
	if err != nil {
		return nil, err
	}
	for newToken.Type != End {
		tokens = append(tokens, newToken)
		newToken, err = l.NextToken()
		if err != nil {
			return nil, err
		}
	}
	return tokens, nil
}
