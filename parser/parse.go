package parser

import (
	"errors"

	"github.com/bogdan-deac/regex/ast"
	"github.com/bogdan-deac/regex/common/generator"
)

type Parser struct {
	lexer    *Lexer
	_mPrefix map[TokenType]PrefixParselet
	_mInfix  map[TokenType]InfixParselet
}

// Pratt parsers are fantastic
// https://journal.stuffwithstuff.com/2011/03/19/pratt-parsers-expression-parsing-made-easy/
func NewParser() *Parser {
	p := &Parser{
		_mPrefix: make(map[TokenType]PrefixParselet),
		_mInfix:  make(map[TokenType]InfixParselet),
		// _mPostfix: make(map[TokenType]PostfixParselet),
	}
	p.RegisterPrefix(Char, CharParselet{})
	p.RegisterPrefix(LParen, GroupParselet{})
	p.RegisterPrefix(Wildcard, WildcardParselet{})

	p.RegisterInfix(Or, OrParselet{})
	p.RegisterInfix(Char, CatParselet{})
	p.RegisterInfix(LParen, CatParselet{})
	p.RegisterInfix(Wildcard, CatParselet{})

	p.RegisterInfix(Star, StarParselet{})
	p.RegisterInfix(Plus, PlusParselet{})
	p.RegisterInfix(Maybe, MaybeParselet{})
	return p
}

func (p *Parser) RegisterPrefix(tt TokenType, pp PrefixParselet) {
	p._mPrefix[tt] = pp
}

func (p *Parser) RegisterInfix(tt TokenType, pp InfixParselet) {
	p._mInfix[tt] = pp
}

func (p *Parser) Parse(regexS string) (ast.Regex[generator.PrintableInt], error) {
	lexer := NewLexer(regexS)
	p.lexer = lexer
	return p.parseExpression()
}

func (p *Parser) parseExpression() (ast.Regex[generator.PrintableInt], error) {
	newTok, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	var leftExpr ast.Regex[generator.PrintableInt]
	parselet, ok := p._mPrefix[newTok.Type]
	if ok {
		leftExpr, err = parselet.Parse(p, newTok)
		if err != nil {
			return nil, err
		}
	}

peek:
	token, err := p.Peek()
	if err != nil {
		return nil, err
	}

	for token.Type != End {
		infixParselet, ok := p._mInfix[token.Type]
		if !ok {
			return leftExpr, nil
		}

		leftExpr, err = infixParselet.Parse(p, leftExpr, token)
		goto peek
	}
	return leftExpr, nil
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
