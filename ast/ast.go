package ast

import (
	"maps"

	"github.com/bogdan-deac/regex/automata"
	"github.com/bogdan-deac/regex/common/generator"
	mapset "github.com/deckarep/golang-set/v2"
)

type Opcode = int

const (
	NoOp Opcode = iota
	CharOp
	CatOp
	OrOp
	StarOp
	PlusOp
	MaybeOp
	WildcardOp
)

//---------------------------
//   Thompson's algorithm
//---------------------------

type Regex[T automata.StateLike] interface {
	Opcode() Opcode
	Optimize() Regex[T]
	Compile(generator.Generator[T]) *automata.NFA[T]
}

type Char[T automata.StateLike] struct {
	Value rune
}

func (Char[T]) Opcode() Opcode { return CharOp }
func (c Char[T]) Compile(gen generator.Generator[T]) *automata.NFA[T] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	return &automata.NFA[T]{
		IntialState: intialState,
		FinalStates: mapset.NewSet(finalState),
		AllStates:   mapset.NewSet(intialState, finalState),
		Alphabet:    mapset.NewSet(c.Value),
		Delta: map[T]map[automata.Symbol][]T{
			intialState: {
				c.Value: {finalState},
			},
		},
		EpsilonTransitions: nil,
	}
}

func (c Char[T]) Optimize() Regex[T] { return c }

type Or[T automata.StateLike] struct {
	Branches []Regex[T]
}

func (Or[T]) Opcode() Opcode { return OrOp }
func (o Or[T]) Compile(gen generator.Generator[T]) *automata.NFA[T] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)
	alphabet := mapset.NewSet[automata.Symbol]()
	epsilonTransitions := make(map[T][]T)
	delta := make(map[T]map[automata.Symbol][]T)
	var branchInitialStates []T

	for _, b := range o.Branches {
		compiledBranch := b.Compile(gen)

		branchInitialStates = append(branchInitialStates, compiledBranch.IntialState)
		// union of the alphabets of the branches
		alphabet = alphabet.Union(compiledBranch.Alphabet)

		// join all states together
		allStates = allStates.Union(compiledBranch.AllStates)

		// the internal epsilon transitions for each branch will remain in the compound automata
		maps.Insert(epsilonTransitions, maps.All(compiledBranch.EpsilonTransitions))

		// for each final state, add an epsilon transition to the new final state
		for fs := range compiledBranch.FinalStates.Iter() {
			epsilonTransitions[fs] = append(epsilonTransitions[fs], finalState)
		}

		// should have no duplicate states, so it's fine to do this
		maps.Insert(delta, maps.All(compiledBranch.Delta))
	}

	// add an epsilon transition from the initial state to all the final states
	epsilonTransitions[intialState] = branchInitialStates

	return &automata.NFA[T]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates,
		Alphabet:           alphabet,
		Delta:              delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (o Or[T]) Optimize() Regex[T] {
	var newBranches []Regex[T]
	for _, b := range o.Branches {
		newBranch := b.Optimize()
		if bo, ok := newBranch.(Or[T]); ok {
			newBranches = append(newBranches, bo.Branches...)
			continue
		}
		newBranches = append(newBranches, newBranch)
	}
	return Or[T]{
		Branches: newBranches,
	}
}

type Star[T automata.StateLike] struct {
	Subexp Regex[T]
}

func (Star[T]) Opcode() Opcode { return StarOp }
func (s Star[T]) Compile(gen generator.Generator[T]) *automata.NFA[T] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)

	subNfa := s.Subexp.Compile(gen)
	epsilonTransitions := maps.Clone(subNfa.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[T][]T)
	}

	epsilonTransitions[intialState] = append(epsilonTransitions[intialState], finalState, subNfa.IntialState)
	for fs := range subNfa.FinalStates.Iter() {
		epsilonTransitions[fs] = append(epsilonTransitions[fs], finalState, subNfa.IntialState)
	}

	return &automata.NFA[T]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates.Union(subNfa.AllStates),
		Alphabet:           subNfa.Alphabet,
		Delta:              subNfa.Delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (s Star[T]) Optimize() Regex[T] { return s }

type Plus[T automata.StateLike] struct {
	Subexp Regex[T]
}

func (Plus[T]) Opcode() Opcode { return PlusOp }

func (p Plus[T]) Compile(gen generator.Generator[T]) *automata.NFA[T] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)

	subNfa := p.Subexp.Compile(gen)
	epsilonTransitions := maps.Clone(subNfa.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[T][]T)
	}

	epsilonTransitions[intialState] = append(epsilonTransitions[intialState], subNfa.IntialState)
	for fs := range subNfa.FinalStates.Iter() {
		epsilonTransitions[fs] = append(epsilonTransitions[fs], finalState, subNfa.IntialState)
	}

	return &automata.NFA[T]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates.Union(subNfa.AllStates),
		Alphabet:           subNfa.Alphabet,
		Delta:              subNfa.Delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (p Plus[T]) Optimize() Regex[T] { return p }

type Cat[T automata.StateLike] struct {
	Left  Regex[T]
	Right Regex[T]
}

func (Cat[T]) Opcode() Opcode { return CatOp }
func (c Cat[T]) Compile(gen generator.Generator[T]) *automata.NFA[T] {
	lc := c.Left.Compile(gen)
	rc := c.Right.Compile(gen)
	allStates := mapset.NewSet[T]().Union(lc.AllStates).Union(rc.AllStates)
	alphabet := mapset.NewSet[automata.Symbol]().Union(lc.Alphabet).Union(rc.Alphabet)

	delta := maps.Clone(lc.Delta)
	maps.Insert(delta, maps.All(rc.Delta))

	epsilonTransitions := maps.Clone(lc.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[T][]T)
	}
	maps.Insert(epsilonTransitions, maps.All(rc.EpsilonTransitions))
	for fs := range lc.FinalStates.Iter() {
		if epsilonTransitions[fs] == nil {
		}
		epsilonTransitions[fs] = append(epsilonTransitions[fs], rc.IntialState)
	}

	return &automata.NFA[T]{
		IntialState:        lc.IntialState,
		FinalStates:        rc.FinalStates,
		AllStates:          allStates,
		Alphabet:           alphabet,
		Delta:              delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (c Cat[T]) Optimize() Regex[T] {
	return Cat[T]{
		Left:  c.Left.Optimize(),
		Right: c.Right.Optimize(),
	}
}

type Maybe[T automata.StateLike] struct {
	Subexp Regex[T]
}

func (Maybe[T]) Opcode() Opcode { return MaybeOp }
func (m Maybe[T]) Compile(gen generator.Generator[T]) *automata.NFA[T] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)

	subNfa := m.Subexp.Compile(gen)
	epsilonTransitions := maps.Clone(subNfa.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[T][]T)
	}

	epsilonTransitions[intialState] = append(epsilonTransitions[intialState], finalState, subNfa.IntialState)
	for fs := range subNfa.FinalStates.Iter() {
		epsilonTransitions[fs] = append(epsilonTransitions[fs], finalState)
	}

	return &automata.NFA[T]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates.Union(subNfa.AllStates),
		Alphabet:           subNfa.Alphabet,
		Delta:              subNfa.Delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (m Maybe[T]) Optimize() Regex[T] { return m }

type Wildcard[T automata.StateLike] struct{}

func (w Wildcard[T]) Opcode() Opcode { return WildcardOp }

func (w Wildcard[T]) Compile(gen generator.Generator[T]) *automata.NFA[T] {
	initialState := gen.Generate()
	finalState := gen.Generate()
	return &automata.NFA[T]{
		IntialState: initialState,
		FinalStates: mapset.NewSet(finalState),
		AllStates:   mapset.NewSet(initialState, finalState),
		Delta: map[T]map[automata.Symbol][]T{
			initialState: {
				automata.Wildcard: {finalState},
			},
		},
		Alphabet: mapset.NewSet[automata.Symbol](automata.Wildcard),
	}
}

func (w Wildcard[T]) Optimize() Regex[T] { return w }
