package parser

import (
	"errors"
	"strconv"

	"github.com/bogdan-deac/regex/ast"
	"github.com/bogdan-deac/regex/common/generator"
)

type Regex = ast.Regex[generator.PrintableInt]

type parser struct {
	index      int
	groupDepth int
}

func NewParser() *parser {
	return &parser{}
}
func (p *parser) Parse(s string) (Regex, error) {
	p.groupDepth = 0
	p.index = 0
	return p.parseAlt(s)
}

func (p *parser) parseStar(s string) bool {
	if len(s) <= p.index {
		return false
	}

	if s[p.index] == '*' {
		return true
	}
	return false
}

func (p *parser) parsePlus(s string) bool {
	if len(s) <= p.index {
		return false
	}
	if s[p.index] == '+' {
		return true
	}
	return false
}

func (p *parser) parseMaybe(s string) bool {
	if len(s) <= p.index {
		return false
	}
	if s[p.index] == '?' {
		return true
	}
	return false
}

func (p *parser) parseQuantifier(s string, atom Regex) (Regex, bool) {
	if p.parseStar(s) {
		return ast.Star[generator.PrintableInt]{Subexp: atom}, true
	}
	if p.parsePlus(s) {
		return ast.Plus[generator.PrintableInt]{Subexp: atom}, true
	}
	if p.parseMaybe(s) {
		return ast.Maybe[generator.PrintableInt]{Subexp: atom}, true
	}

	return nil, false
}

func (p *parser) parseRepeat(s string) (Regex, error) {
	atom, err := p.parseAtom(s)
	if err != nil {
		return nil, err
	}
	if quantifiedAtom, ok := p.parseQuantifier(s, atom); ok {
		p.index++
		return quantifiedAtom, nil
	}

	return atom, nil
}

func (p *parser) parseConcat(s string) (Regex, error) {
	regex, err := p.parseRepeat(s)
	if err != nil {
		return nil, err
	}
	for {
		newRegex, err := p.parseRepeat(s)
		if err != nil {
			return nil, err
		}
		if newRegex == nil {
			return regex, nil
		}
		regex = ast.Cat[generator.PrintableInt]{
			Left:  regex,
			Right: newRegex,
		}
	}
}

func (p *parser) parseAlt(s string) (Regex, error) {
	regex, err := p.parseConcat(s)
	if err != nil {
		return nil, err

	}
	for p.index < len(s) && s[p.index] == '|' {
		p.index++
		newRegex, err := p.parseConcat(s)
		if err != nil {
			return nil, err
		}
		regex = ast.Or[generator.PrintableInt]{
			Branches: []Regex{
				regex,
				newRegex,
			},
		}
	}
	return regex, nil
}

func (p *parser) parseGroup(s string) (Regex, error) {
	if p.index < len(s) && s[p.index] == '(' {
		p.groupDepth++
		p.index++
		regex, err := p.parseAlt(s)
		if err != nil {
			return nil, err
		}
		if p.index < len(s) && s[p.index] == ')' {
			p.index++
			p.groupDepth--
			return regex, nil
		}
		return nil, errors.New("expected closing bracket but found none at index " + strconv.Itoa(p.index))
	}
	return nil, nil
}

func (p *parser) parseLiteral(s string) (Regex, error) {
	if len(s) <= p.index {
		return nil, nil
	}

	switch s[p.index] {
	case '*', '+', '?':
		return nil, errors.New("found unexpected operator at index " + strconv.Itoa(p.index))
	case '|':
		return nil, nil
	case '(':
		return nil, nil
	case ')':
		if p.groupDepth == 0 {
			return nil, errors.New("found unexpected closing paren at index " + strconv.Itoa(p.index))
		}
		return nil, nil
	case '.':
		return ast.Wildcard[generator.PrintableInt]{}, nil
	case '\\':
		p.index++
		if len(s) <= p.index {
			return nil, errors.New("found escape operator without argument at index" + strconv.Itoa(p.index))
		}
		fallthrough
	default:

		return ast.Char[generator.PrintableInt]{
			// TBD unicode suport
			Value: rune(s[p.index]),
		}, nil
	}
}
func (p *parser) parseAtom(s string) (Regex, error) {
	// attempt parsing a literal
	regex, err := p.parseLiteral(s)
	if err != nil {
		return nil, err
	}
	if regex != nil {
		p.index++
		return regex, nil
	}

	// otherwise, a group
	regex, err = p.parseGroup(s)
	if err != nil {
		return nil, err
	}

	return regex, nil
}
