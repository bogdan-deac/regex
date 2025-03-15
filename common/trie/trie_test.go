package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrie(t *testing.T) {
	trie := NewTrie[int, int]()
	trie.Insert([]int{1, 2, 3}, 4)
	v := trie.Lookup([]int{1, 2, 3})
	assert.Equal(t, *v, 4)

	trie.Insert([]int{2, 3, 5}, 7)
	v = trie.Lookup([]int{2, 3, 5})
	assert.Equal(t, *v, 7)
}
