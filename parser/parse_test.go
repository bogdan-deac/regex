package parser

import (
	"testing"

	"github.com/bogdan-deac/regex"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tt := []struct {
		reS            string
		expectedResult regex.Regex
	}{
		{
			reS:            "a",
			expectedResult: regex.Char{Value: 'a'},
		},
		{
			reS: "ab",
			expectedResult: regex.Cat{
				Left:  regex.Char{Value: 'a'},
				Right: regex.Char{Value: 'b'},
			},
		},
		{
			reS: "a|b",
			expectedResult: regex.Or{
				Branches: []regex.Regex{
					regex.Char{Value: 'a'},
					regex.Char{Value: 'b'},
				},
			},
		},
		{
			reS: "a*",
			expectedResult: regex.Star{
				Subexp: regex.Char{Value: 'a'},
			},
		},
		{
			reS: "ab*",
			expectedResult: regex.Cat{
				Left: regex.Char{Value: 'a'},
				Right: regex.Star{
					Subexp: regex.Char{Value: 'b'},
				},
			},
		},
		{
			reS:            "(a)",
			expectedResult: regex.Char{Value: 'a'},
		},
		{
			reS: "(a|b)",
			expectedResult: regex.Or{
				Branches: []regex.Regex{
					regex.Char{Value: 'a'},
					regex.Char{Value: 'b'},
				},
			},
		},
		{
			reS: "(a|b)*",
			expectedResult: regex.Star{
				Subexp: regex.Or{
					Branches: []regex.Regex{
						regex.Char{Value: 'a'},
						regex.Char{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a*|b",
			expectedResult: regex.Or{
				Branches: []regex.Regex{
					regex.Star{
						Subexp: regex.Char{Value: 'a'},
					},
					regex.Char{Value: 'b'},
				},
			},
		},
		{
			reS: "a|b*",
			expectedResult: regex.Or{
				Branches: []regex.Regex{
					regex.Char{Value: 'a'},
					regex.Star{
						Subexp: regex.Char{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a?",
			expectedResult: regex.Maybe{
				Subexp: regex.Char{Value: 'a'},
			},
		},
		{
			reS: "a*|b?",
			expectedResult: regex.Or{
				Branches: []regex.Regex{
					regex.Star{
						Subexp: regex.Char{Value: 'a'},
					},
					regex.Maybe{
						Subexp: regex.Char{Value: 'b'},
					},
				},
			},
		},
		{
			reS: "a*b",
			expectedResult: regex.Cat{
				Left: regex.Star{
					Subexp: regex.Char{Value: 'a'},
				},
				Right: regex.Char{Value: 'b'},
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
