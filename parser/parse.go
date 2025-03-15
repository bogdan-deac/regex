package parser

import (
	"errors"

	"github.com/bogdan-deac/regex"
)

type Parser struct {
	lexer     *Lexer
	_mPrefix  map[TokenType]PrefixParselet
	_mInfix   map[TokenType]InfixParselet
	_mPostfix map[TokenType]PostfixParselet
}

// Pratt parsers are fantastic
// https://journal.stuffwithstuff.com/2011/03/19/pratt-parsers-expression-parsing-made-easy/
func NewParser() *Parser {
	p := &Parser{
		_mPrefix:  make(map[TokenType]PrefixParselet),
		_mInfix:   make(map[TokenType]InfixParselet),
		_mPostfix: make(map[TokenType]PostfixParselet),
	}
	p.RegisterPrefix(Char, CharParselet{})
	p.RegisterPrefix(LParen, GroupParselet{})

	p.RegisterInfix(Or, OrParselet{})
	p.RegisterInfix(Char, CatParselet{})
	p.RegisterInfix(LParen, CatParselet{})

	p.RegisterPostfix(Star, StarParselet{})
	p.RegisterPostfix(Maybe, MaybeParselet{})
	return p
}

func (p *Parser) RegisterPrefix(tt TokenType, pp PrefixParselet) {
	p._mPrefix[tt] = pp
}

func (p *Parser) RegisterInfix(tt TokenType, pp InfixParselet) {
	p._mInfix[tt] = pp
}

func (p *Parser) RegisterPostfix(tt TokenType, pp InfixParselet) {
	p._mPostfix[tt] = pp
}

func (p *Parser) Parse(regexS string) (regex.Regex, error) {
	lexer := NewLexer(regexS)
	p.lexer = lexer
	return p.parseExpression()
}
func (p *Parser) parseExpression() (regex.Regex, error) {
	newTok, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	var leftExpr regex.Regex
	parselet, ok := p._mPrefix[newTok.Type]
	if ok {
		leftExpr, err = parselet.Parse(p, newTok)
		if err != nil {
			return nil, err
		}
	}

	token, err := p.Peek()
	if err != nil {
		return nil, err
	}
	pfParselet, ok := p._mPostfix[token.Type]
	if ok {
		leftExpr, err = pfParselet.Parse(p, leftExpr, newTok)
		if err != nil {
			return nil, err
		}
		token, err = p.Peek()
		if err != nil {
			return nil, err
		}
	}

	infixParselet, ok := p._mInfix[token.Type]
	if !ok {
		return leftExpr, nil
	}

	expr, err := infixParselet.Parse(p, leftExpr, token)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) Peek() (Token, error) {
	if p.lexer.pos >= len(p.lexer.input) {
		return Token{Type: End}, nil
	}
	return p.lexer.Peek()
}

func (p *Parser) Consume() {
	p.lexer.pos++
}

func (p *Parser) ConsumeToken(t Token) error {
	tok, err := p.Peek()
	if err != nil {
		return err
	}
	if tok.Type != t.Type {
		return errors.New("Got different token than expected " + string(t.Value))
	}
	p.Consume()
	return nil
}
