package automata

import (
	"cmp"

	set "github.com/deckarep/golang-set/v2"
)

type DFA[T cmp.Ordered] struct {
	IntialState T
	FinalStates set.Set[T]
	AllStates   set.Set[T]
	Delta       map[T]map[Symbol]T
	SinkState   *T
	Alphabet    set.Set[Symbol]
}

func NewDFA[T cmp.Ordered](
	IntialState T,
	FinalStates set.Set[T],
	AllStates set.Set[T],
	Delta map[T]map[Symbol]T,
	SinkState *T,
) *DFA[T] {
	return &DFA[T]{
		IntialState: IntialState,
		FinalStates: FinalStates,
		AllStates:   AllStates,
		Delta:       Delta,
		SinkState:   SinkState,
	}
}

func (dfa *DFA[T]) Accepts(input []Symbol) bool {
	currentState := dfa.IntialState
	for _, symbol := range input {
		if _, ok := dfa.Delta[currentState][symbol]; !ok {
			return false
		}
		currentState = dfa.Delta[currentState][symbol]
		// if any transition leads to the sink state, break early - no match
		if dfa.SinkState != nil && currentState == *dfa.SinkState {
			return false
		}
	}
	return dfa.FinalStates.Contains(currentState)
}
