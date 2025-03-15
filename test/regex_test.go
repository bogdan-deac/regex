package regex_test

import (
	"testing"

	"github.com/bogdan-deac/regex/automata"
	"github.com/bogdan-deac/regex/common/generator"
	"github.com/bogdan-deac/regex/parser"
	"github.com/stretchr/testify/assert"
)

func TestRegex(t *testing.T) {
	tt := []struct {
		regexS     string
		mustAccept []string
	}{
		{
			regexS:     "a",
			mustAccept: []string{"a"},
		},
		{
			regexS:     "ab",
			mustAccept: []string{"ab"},
		},
		{
			regexS:     "a*",
			mustAccept: []string{"", "a", "aa", "aaa", "aaaa", "aaaaaaaaaaaaaaaa"},
		},
		{
			regexS:     "a|b",
			mustAccept: []string{"a", "b"},
		},
		{
			regexS:     "(a|b)c",
			mustAccept: []string{"ac", "bc"},
		},
		{
			regexS:     "(a|b)*c",
			mustAccept: []string{"c", "ac", "abbac", "abbbbc", "bbbbbc"},
		},
		{
			regexS:     "a?(b|c)",
			mustAccept: []string{"b", "c", "ab", "ac"},
		},
		{
			regexS:     "a?|b*",
			mustAccept: []string{"", "a", "b", "bb"},
		},
		{
			regexS:     "(a|b)?c*",
			mustAccept: []string{"", "a", "b", "c", "ac", "bc", "cc", "acc", "bcc", "ccc"},
		},
		{
			regexS:     "a|b|c",
			mustAccept: []string{"a", "b", "c"},
		},
		{
			regexS:     "aa?",
			mustAccept: []string{"a", "aa"},
		},
	}
	p := parser.NewParser()
	for _, tc := range tt {
		regex, err := p.Parse(tc.regexS)
		assert.Nil(t, err)
		g := generator.NewIntGenerator()
		dfa := regex.Compile(g).ToDFA(g)
		for _, s := range tc.mustAccept {
			assert.True(t, dfa.Accepts([]automata.Symbol(s)))
		}
	}
}
