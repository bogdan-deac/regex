package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	tt := []struct {
		input          string
		expectedTokens []Token
	}{
		{
			input: "a",
			expectedTokens: []Token{
				{Type: Char, Value: 'a'},
			},
		},
		{
			input: "a*",
			expectedTokens: []Token{
				{Type: Char, Value: 'a'},
				{Type: Star, Value: '*'},
			},
		},
		{
			input: "a|b*",
			expectedTokens: []Token{
				{Type: Char, Value: 'a'},
				{Type: Or, Value: '|'},
				{Type: Char, Value: 'b'},
				{Type: Star, Value: '*'},
			},
		},
		{
			input: "()(a|b*)",
			expectedTokens: []Token{
				{Type: LParen, Value: '('},
				{Type: RParen, Value: ')'},
				{Type: LParen, Value: '('},
				{Type: Char, Value: 'a'},
				{Type: Or, Value: '|'},
				{Type: Char, Value: 'b'},
				{Type: Star, Value: '*'},
				{Type: RParen, Value: ')'},
			},
		},
		{
			input: "\\(",
			expectedTokens: []Token{
				{Type: Char, Value: '('},
			},
		},
	}
	for _, tc := range tt {
		l := NewLexer(tc.input)
		tokens, err := l.Tokenize()
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedTokens, tokens)
	}
}
