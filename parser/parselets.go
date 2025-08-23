package parser

import (
	"errors"

	"github.com/bogdan-deac/regex/ast"
	"github.com/bogdan-deac/regex/common/generator"
)

// ------------------------
//
//	PREFIX PARSELETS
//
// ------------------------
type PrefixParselet interface {
	Parse(*Parser, Token) (ast.Regex[generator.PrintableInt], error)
}
type CharParselet struct{}

func (CharParselet) Parse(p *Parser, t Token) (ast.Regex[generator.PrintableInt], error) {
	return ast.Char[generator.PrintableInt]{Value: t.Value}, nil
}

type WildcardParselet struct{}

func (WildcardParselet) Parse(p *Parser, t Token) (ast.Regex[generator.PrintableInt], error) {
	return ast.Wildcard[generator.PrintableInt]{}, nil
}

type GroupParselet struct{}

func (GroupParselet) Parse(p *Parser, t Token) (ast.Regex[generator.PrintableInt], error) {
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
	Parse(*Parser, ast.Regex[generator.PrintableInt], Token) (ast.Regex[generator.PrintableInt], error)
}

type OrParselet struct{}

func (OrParselet) Parse(p *Parser, left ast.Regex[generator.PrintableInt], t Token) (ast.Regex[generator.PrintableInt], error) {
	_ = p.ConsumeToken(mkOpToken(Or))
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return ast.Or[generator.PrintableInt]{
		Branches: []ast.Regex[generator.PrintableInt]{left, right},
	}, nil
}

type CatParselet struct{}

func (CatParselet) Parse(p *Parser, left ast.Regex[generator.PrintableInt], t Token) (ast.Regex[generator.PrintableInt], error) {
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return ast.Cat[generator.PrintableInt]{
		Left:  left,
		Right: right,
	}, nil
}

type StarParselet struct{}

func (StarParselet) Parse(p *Parser, left ast.Regex[generator.PrintableInt], t Token) (ast.Regex[generator.PrintableInt], error) {
	_ = p.ConsumeToken(mkOpToken(Star))
	if left == nil {
		return nil, errors.New("Detected star operator without argument")
	}
	return ast.Star[generator.PrintableInt]{Subexp: left}, nil
}

type PlusParselet struct{}

func (PlusParselet) Parse(p *Parser, left ast.Regex[generator.PrintableInt], t Token) (ast.Regex[generator.PrintableInt], error) {
	_ = p.ConsumeToken(mkOpToken(Plus))
	if left == nil {
		return nil, errors.New("Detected plus operator without argument")
	}
	return ast.Plus[generator.PrintableInt]{Subexp: left}, nil
}

type MaybeParselet struct{}

func (MaybeParselet) Parse(p *Parser, left ast.Regex[generator.PrintableInt], t Token) (ast.Regex[generator.PrintableInt], error) {
	_ = p.ConsumeToken(mkOpToken(Maybe))
	if left == nil {
		return nil, errors.New("Detected maybe operator without argument")
	}
	return ast.Maybe[generator.PrintableInt]{Subexp: left}, nil
}
