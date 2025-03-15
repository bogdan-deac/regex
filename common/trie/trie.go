package trie

import "errors"

type Trie[T comparable, V any] struct {
	Children map[T]*Trie[T, V]
	Value    *V
}

func NewTrie[T comparable, V any]() *Trie[T, V] {
	return &Trie[T, V]{}
}

func (t *Trie[T, V]) Insert(elems []T, value V) error {
	crt := t
	elIndex := 0
	for elIndex < len(elems) {
		if len(crt.Children) == 0 {
			crt.Children = make(map[T]*Trie[T, V])
		}
		child, ok := crt.Children[elems[elIndex]]
		if !ok {
			child = NewTrie[T, V]()
			crt.Children[elems[elIndex]] = child
		}
		crt = child
		elIndex++
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
