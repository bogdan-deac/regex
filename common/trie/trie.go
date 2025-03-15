package trie

import "errors"

type Trie[T comparable, V any] struct {
	Children map[T]*Trie[T, V]
	Value    *V
}

func NewTrie[T comparable, V any]() *Trie[T, V] {
	return &Trie[T, V]{
		Children: make(map[T]*Trie[T, V]),
	}
}

func (t *Trie[T, V]) Insert(elems []T, value V) error {
	crt := t

	for _, elem := range elems {
		child, ok := crt.Children[elem]
		if !ok {
			child = NewTrie[T, V]()
			crt.Children[elem] = child
		}
		crt = child
	}
	if crt.Value != nil {
		return errors.New("Value already exists")
	}
	crt.Value = &value
	return nil
}

func (t *Trie[T, V]) Lookup(elems []T) *V {
	crt := t

	for e := range elems {
		if len(crt.Children) == 0 {
			return nil
		}
		child, ok := crt.Children[elems[e]]
		if !ok {
			return nil
		}
		crt = child
	}
	return crt.Value
}
