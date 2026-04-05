package parser

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/bogdan-deac/regex/ast"
	"github.com/bogdan-deac/regex/automata"
	"github.com/bogdan-deac/regex/common/generator"
)

type Regex = ast.Regex[generator.PrintableInt]

type parser struct {
	index      int
	groupDepth int
	inSet      bool
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
	case '|', '(', '[':
		return nil, nil
	case ')':
		if p.groupDepth == 0 {
			return nil, errors.New("found unexpected closing paren at index " + strconv.Itoa(p.index))
		}
		return nil, nil
	case ']':
		if !p.inSet {
			return nil, errors.New("found unexpected closing square bracket at index " + strconv.Itoa(p.index))
		}
		return nil, nil
	case '.':
		p.index++
		return ast.Wildcard[generator.PrintableInt]{}, nil
	case '\\':
		if len(s) <= p.index {
			return nil, errors.New("found escape operator without argument at index" + strconv.Itoa(p.index))
		}

		p.index++
		fallthrough
	default:
		val := rune(s[p.index])
		p.index++
		return ast.Char[generator.PrintableInt]{
			// TBD unicode suport
			Value: val,
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
		return regex, nil
	}

	// otherwise, a group
	regex, err = p.parseGroup(s)
	if err != nil {
		return nil, err
	}
	if regex != nil {
		return regex, nil
	}

	regex, err = p.parseSet(s)
	if err != nil {
		return nil, err
	}

	return regex, nil
}
func (p *parser) parseSet(s string) (Regex, error) {

	if p.index < len(s) && s[p.index] != '[' || p.index == len(s) {
		return nil, nil
	}
	p.index++
	if p.index < len(s) && s[p.index] == ']' {
		return nil, errors.New("Detected empty character set in regex: []")
	}
	rangesToAdd := make([][]rune, 0)
	rangesToRemove := make([][]rune, 0)
	literalsToAdd := make([]rune, 0)
	literalsToRemove := make([]rune, 0)
	isNegated := false
	if p.index < len(s) && s[p.index] == '^' {
		isNegated = true
		p.index++
		if p.index < len(s) && s[p.index] == ']' {
			return nil, errors.New("Detected useless negation in the character set")
		}
	}
	for p.index < len(s) && s[p.index] != ']' {
		if p.index < len(s) && s[p.index] == '\\' {
			p.index++
		}
		if p.index+1 < len(s) && s[p.index+1] == '-' {
			// we have a range
			if p.index+2 >= len(s) {
				return nil, errors.New("Detected unclosed range at the end of the charater group")
			}
			rangeStart := s[p.index]
			rangeEnd := s[p.index+2]
			if rangeStart > rangeEnd {
				return nil, fmt.Errorf("Detected invalid range between %q and %q", rangeStart, rangeEnd)
			}
			if isNegated {
				rangesToRemove = append(rangesToRemove, automata.ASCIISet{rune(rangeStart), rune(rangeEnd)})
			} else {
				rangesToAdd = append(rangesToAdd, automata.ASCIISet{rune(rangeStart), rune(rangeEnd)})
			}
			p.index += 3
			continue
		}

		if isNegated {
			literalsToRemove = append(literalsToRemove, rune(s[p.index]))
		} else {
			literalsToAdd = append(literalsToAdd, rune(s[p.index]))
		}
		p.index++

	}
	if s[p.index] != ']' {
		return nil, errors.New("Detected unclosed character set")
	}
	p.index++

	finalRage := automata.ASCIISet{}

	if isNegated {
		finalRage = slices.Clone(automata.ASCIIChars)
		for _, rangeToRemove := range rangesToRemove {
			finalRage = finalRage.Difference(rangeToRemove[0], rangeToRemove[1])
		}
		for _, literalToRemove := range literalsToRemove {
			finalRage = finalRage.Delete(literalToRemove)
		}
	} else {
		for _, rangeToAdd := range rangesToAdd {
			finalRage = finalRage.Union(rangeToAdd[0], rangeToAdd[1])
		}
		for _, literalToAdd := range literalsToAdd {
			finalRage = finalRage.Add(literalToAdd)
		}
	}

	orExp := ast.Or[generator.PrintableInt]{
		Branches: []Regex{},
	}
	for _, c := range finalRage {
		orExp.Branches = append(orExp.Branches, ast.Char[generator.PrintableInt]{
			Value: c,
		})
	}
	return orExp, nil
}
