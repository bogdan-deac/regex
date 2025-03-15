package parser

import (
	"errors"

	"github.com/bogdan-deac/regex"
)

// ------------------------
//
//	PREFIX PARSELETS
//
// ------------------------
type PrefixParselet interface {
	Parse(*Parser, Token) (regex.Regex, error)
}
type CharParselet struct{}

func (CharParselet) Parse(p *Parser, t Token) (regex.Regex, error) {
	return regex.Char{Value: t.Value}, nil
}

type GroupParselet struct{}

func (GroupParselet) Parse(p *Parser, t Token) (regex.Regex, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.ConsumeToken(mkOpToken(RParen))
	return expr, nil
}

// ------------------------
//     INFIX PARSELETS
// ------------------------

type InfixParselet interface {
	Parse(*Parser, regex.Regex, Token) (regex.Regex, error)
}

type OrParselet struct{}

func (OrParselet) Parse(p *Parser, left regex.Regex, t Token) (regex.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Or))
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return regex.Or{
		Branches: []regex.Regex{left, right},
	}, nil
}

type CatParselet struct{}

func (CatParselet) Parse(p *Parser, left regex.Regex, t Token) (regex.Regex, error) {
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return regex.Cat{
		Left:  left,
		Right: right,
	}, nil
}

// ------------------------
//     POSTFIX PARSELETS
// ------------------------

type PostfixParselet interface {
	Parse(*Parser, regex.Regex, Token) (regex.Regex, error)
}

type StarParselet struct{}

func (StarParselet) Parse(p *Parser, left regex.Regex, t Token) (regex.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Star))
	if left == nil {
		return nil, errors.New("Detected star operator without argument")
	}
	return regex.Star{Subexp: left}, nil
}

type PlusParselet struct{}

func (PlusParselet) Parse(p *Parser, left regex.Regex, t Token) (regex.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Plus))
	if left == nil {
		return nil, errors.New("Detected plus operator without argument")
	}
	return regex.Plus{Subexp: left}, nil
}

type MaybeParselet struct{}

func (MaybeParselet) Parse(p *Parser, left regex.Regex, t Token) (regex.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Maybe))
	if left == nil {
		return nil, errors.New("Detected maybe operator without argument")
	}
	return regex.Maybe{Subexp: left}, nil
}
