package parser

import (
	"errors"

	"github.com/bogdan-deac/regex/ast"
)

// ------------------------
//
//	PREFIX PARSELETS
//
// ------------------------
type PrefixParselet interface {
	Parse(*Parser, Token) (ast.Regex, error)
}
type CharParselet struct{}

func (CharParselet) Parse(p *Parser, t Token) (ast.Regex, error) {
	return ast.Char{Value: t.Value}, nil
}

type GroupParselet struct{}

func (GroupParselet) Parse(p *Parser, t Token) (ast.Regex, error) {
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
	Parse(*Parser, ast.Regex, Token) (ast.Regex, error)
}

type OrParselet struct{}

func (OrParselet) Parse(p *Parser, left ast.Regex, t Token) (ast.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Or))
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return ast.Or{
		Branches: []ast.Regex{left, right},
	}, nil
}

type CatParselet struct{}

func (CatParselet) Parse(p *Parser, left ast.Regex, t Token) (ast.Regex, error) {
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return ast.Cat{
		Left:  left,
		Right: right,
	}, nil
}

// ------------------------
//     POSTFIX PARSELETS
// ------------------------

type PostfixParselet interface {
	Parse(*Parser, ast.Regex, Token) (ast.Regex, error)
}

type StarParselet struct{}

func (StarParselet) Parse(p *Parser, left ast.Regex, t Token) (ast.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Star))
	if left == nil {
		return nil, errors.New("Detected star operator without argument")
	}
	return ast.Star{Subexp: left}, nil
}

type PlusParselet struct{}

func (PlusParselet) Parse(p *Parser, left ast.Regex, t Token) (ast.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Plus))
	if left == nil {
		return nil, errors.New("Detected plus operator without argument")
	}
	return ast.Plus{Subexp: left}, nil
}

type MaybeParselet struct{}

func (MaybeParselet) Parse(p *Parser, left ast.Regex, t Token) (ast.Regex, error) {
	_ = p.ConsumeToken(mkOpToken(Maybe))
	if left == nil {
		return nil, errors.New("Detected maybe operator without argument")
	}
	return ast.Maybe{Subexp: left}, nil
}
