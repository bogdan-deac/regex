package parser

import (
	"testing"

	"github.com/bogdan-deac/regex/ast"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tt := []struct {
		reS            string
		expectedResult ast.Regex
	}{
		{
			reS:            "a",
			expectedResult: ast.Char{Value: 'a'},
		},
		{
			reS: "ab",
			expectedResult: ast.Cat{
				Left:  ast.Char{Value: 'a'},
				Right: ast.Char{Value: 'b'},
			},
		},
		{
			reS: "a|b",
			expectedResult: ast.Or{
				Branches: []ast.Regex{
					ast.Char{Value: 'a'},
					ast.Char{Value: 'b'},
				},
			},
		},
		{
			reS: "a*",
			expectedResult: ast.Star{
				Subexp: ast.Char{Value: 'a'},
			},
		},
		{
			reS: "ab*",
			expectedResult: ast.Cat{
				Left: ast.Char{Value: 'a'},
				Right: ast.Star{
					Subexp: ast.Char{Value: 'b'},
				},
			},
		},
		{
			reS:            "(a)",
			expectedResult: ast.Char{Value: 'a'},
		},
		{
			reS: "(a|b)",
			expectedResult: ast.Or{
				Branches: []ast.Regex{
					ast.Char{Value: 'a'},
					ast.Char{Value: 'b'},
				},
			},
		},
		{
			reS: "(a|b)*",
			expectedResult: ast.Star{
				Subexp: ast.Or{
					Branches: []ast.Regex{
						ast.Char{Value: 'a'},
						ast.Char{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a*|b",
			expectedResult: ast.Or{
				Branches: []ast.Regex{
					ast.Star{
						Subexp: ast.Char{Value: 'a'},
					},
					ast.Char{Value: 'b'},
				},
			},
		},
		{
			reS: "a|b*",
			expectedResult: ast.Or{
				Branches: []ast.Regex{
					ast.Char{Value: 'a'},
					ast.Star{
						Subexp: ast.Char{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a?",
			expectedResult: ast.Maybe{
				Subexp: ast.Char{Value: 'a'},
			},
		},
		{
			reS: "a*|b?",
			expectedResult: ast.Or{
				Branches: []ast.Regex{
					ast.Star{
						Subexp: ast.Char{Value: 'a'},
					},
					ast.Maybe{
						Subexp: ast.Char{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a*b",
			expectedResult: ast.Cat{
				Left: ast.Star{
					Subexp: ast.Char{Value: 'a'},
				},
				Right: ast.Char{Value: 'b'},
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
