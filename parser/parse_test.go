package parser

import (
	"testing"

	"github.com/bogdan-deac/regex/ast"
	"github.com/bogdan-deac/regex/common/generator"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tt := []struct {
		reS            string
		expectedResult ast.Regex[generator.PrintableInt]
	}{
		{
			reS:            "a",
			expectedResult: ast.Char[generator.PrintableInt]{Value: 'a'},
		},
		{
			reS: "ab",
			expectedResult: ast.Cat[generator.PrintableInt]{
				Left:  ast.Char[generator.PrintableInt]{Value: 'a'},
				Right: ast.Char[generator.PrintableInt]{Value: 'b'},
			},
		},
		{
			reS: "a|b",
			expectedResult: ast.Or[generator.PrintableInt]{
				Branches: []ast.Regex[generator.PrintableInt]{
					ast.Char[generator.PrintableInt]{Value: 'a'},
					ast.Char[generator.PrintableInt]{Value: 'b'},
				},
			},
		},
		{
			reS: "a*",
			expectedResult: ast.Star[generator.PrintableInt]{
				Subexp: ast.Char[generator.PrintableInt]{Value: 'a'},
			},
		},
		{
			reS: "ab*",
			expectedResult: ast.Cat[generator.PrintableInt]{
				Left: ast.Char[generator.PrintableInt]{Value: 'a'},
				Right: ast.Star[generator.PrintableInt]{
					Subexp: ast.Char[generator.PrintableInt]{Value: 'b'},
				},
			},
		},
		{
			reS:            "(a)",
			expectedResult: ast.Char[generator.PrintableInt]{Value: 'a'},
		},
		{
			reS: "(a|b)",
			expectedResult: ast.Or[generator.PrintableInt]{
				Branches: []ast.Regex[generator.PrintableInt]{
					ast.Char[generator.PrintableInt]{Value: 'a'},
					ast.Char[generator.PrintableInt]{Value: 'b'},
				},
			},
		},
		{
			reS: "(a|b)*",
			expectedResult: ast.Star[generator.PrintableInt]{
				Subexp: ast.Or[generator.PrintableInt]{
					Branches: []ast.Regex[generator.PrintableInt]{
						ast.Char[generator.PrintableInt]{Value: 'a'},
						ast.Char[generator.PrintableInt]{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a*|b",
			expectedResult: ast.Or[generator.PrintableInt]{
				Branches: []ast.Regex[generator.PrintableInt]{
					ast.Star[generator.PrintableInt]{
						Subexp: ast.Char[generator.PrintableInt]{Value: 'a'},
					},
					ast.Char[generator.PrintableInt]{Value: 'b'},
				},
			},
		},
		{
			reS: "a|b*",
			expectedResult: ast.Or[generator.PrintableInt]{
				Branches: []ast.Regex[generator.PrintableInt]{
					ast.Char[generator.PrintableInt]{Value: 'a'},
					ast.Star[generator.PrintableInt]{
						Subexp: ast.Char[generator.PrintableInt]{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a?",
			expectedResult: ast.Maybe[generator.PrintableInt]{
				Subexp: ast.Char[generator.PrintableInt]{Value: 'a'},
			},
		},
		{
			reS: "a*|b?",
			expectedResult: ast.Or[generator.PrintableInt]{
				Branches: []ast.Regex[generator.PrintableInt]{
					ast.Star[generator.PrintableInt]{
						Subexp: ast.Char[generator.PrintableInt]{Value: 'a'},
					},
					ast.Maybe[generator.PrintableInt]{
						Subexp: ast.Char[generator.PrintableInt]{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a*b",
			expectedResult: ast.Cat[generator.PrintableInt]{
				Left: ast.Star[generator.PrintableInt]{
					Subexp: ast.Char[generator.PrintableInt]{Value: 'a'},
				},
				Right: ast.Char[generator.PrintableInt]{Value: 'b'},
			},
		},
		{
			reS: "(a|b|c)",
			expectedResult: ast.Or[generator.PrintableInt]{
				Branches: []ast.Regex[generator.PrintableInt]{
					ast.Or[generator.PrintableInt]{
						Branches: []ast.Regex[generator.PrintableInt]{
							ast.Char[generator.PrintableInt]{Value: 'a'},
							ast.Char[generator.PrintableInt]{Value: 'b'},
						},
					},
					ast.Char[generator.PrintableInt]{Value: 'c'},
				},
			},
		},
		{
			reS: "(aa|bb|cc)",
			expectedResult: ast.Or[generator.PrintableInt]{
				Branches: []ast.Regex[generator.PrintableInt]{
					ast.Or[generator.PrintableInt]{
						Branches: []ast.Regex[generator.PrintableInt]{
							ast.Cat[generator.PrintableInt]{
								Left:  ast.Char[generator.PrintableInt]{Value: 'a'},
								Right: ast.Char[generator.PrintableInt]{Value: 'a'},
							},
							ast.Cat[generator.PrintableInt]{
								Left:  ast.Char[generator.PrintableInt]{Value: 'b'},
								Right: ast.Char[generator.PrintableInt]{Value: 'b'},
							},
						},
					},
					ast.Cat[generator.PrintableInt]{
						Left:  ast.Char[generator.PrintableInt]{Value: 'c'},
						Right: ast.Char[generator.PrintableInt]{Value: 'c'},
					},
				},
			},
		},
	}

	p := NewParser()
	for _, tc := range tt {
		exp, err := p.Parse(tc.reS)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedResult, exp)
	}
}
