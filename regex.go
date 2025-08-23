package regex

import (
	"github.com/bogdan-deac/regex/parser"

	"github.com/bogdan-deac/regex/automata"
	"github.com/bogdan-deac/regex/common/generator"
)

func Compile(reS string) (*automata.DFA[generator.PrintableInt], error) {
	g := generator.NewIntGenerator()
	p := parser.NewParser()
	re, err := p.Parse(reS)
	if err != nil {
		return nil, err
	}
	re = re.Optimize()
	return re.Compile(g).ToDFA(g).Minimize(), nil
}
