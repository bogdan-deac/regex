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
)

//---------------------------
//   Thompson's algorithm
//---------------------------

type Regex interface {
	Opcode() Opcode
	Optimize() Regex
	Compile(generator.Generator[int]) *automata.NFA[int]
}

type Char struct {
	Value rune
}

func (Char) Opcode() Opcode { return CharOp }
func (c Char) Compile(gen generator.Generator[int]) *automata.NFA[int] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	return &automata.NFA[int]{
		IntialState: intialState,
		FinalStates: mapset.NewSet(finalState),
		AllStates:   mapset.NewSet(intialState, finalState),
		Alphabet:    mapset.NewSet(c.Value),
		Delta: map[int]map[automata.Symbol][]int{
			intialState: {
				c.Value: {finalState},
			},
		},
		EpsilonTransitions: nil,
	}
}

func (c Char) Optimize() Regex { return c }

type Or struct {
	Branches []Regex
}

func (Or) Opcode() Opcode { return OrOp }
func (o Or) Compile(gen generator.Generator[int]) *automata.NFA[int] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)
	alphabet := mapset.NewSet[automata.Symbol]()
	epsilonTransitions := make(map[int][]int)
	delta := make(map[int]map[automata.Symbol][]int)
	var branchInitialStates []int

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

	return &automata.NFA[int]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates,
		Alphabet:           alphabet,
		Delta:              delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (o Or) Optimize() Regex {
	var newBranches []Regex
	for _, b := range o.Branches {
		newBranch := b.Optimize()
		if bo, ok := newBranch.(Or); ok {
			newBranches = append(newBranches, bo.Branches...)
			continue
		}
		newBranches = append(newBranches, newBranch)
	}
	return Or{
		Branches: newBranches,
	}
}

type Star struct {
	Subexp Regex
}

func (Star) Opcode() Opcode { return StarOp }
func (s Star) Compile(gen generator.Generator[int]) *automata.NFA[int] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)

	subNfa := s.Subexp.Compile(gen)
	epsilonTransitions := maps.Clone(subNfa.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[int][]int)
	}

	epsilonTransitions[intialState] = append(epsilonTransitions[intialState], finalState, subNfa.IntialState)
	for fs := range subNfa.FinalStates.Iter() {
		epsilonTransitions[fs] = append(epsilonTransitions[fs], finalState, subNfa.IntialState)
	}

	return &automata.NFA[int]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates.Union(subNfa.AllStates),
		Alphabet:           subNfa.Alphabet,
		Delta:              subNfa.Delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (s Star) Optimize() Regex { return s }

type Plus struct {
	Subexp Regex
}

func (Plus) Opcode() Opcode { return PlusOp }

func (p Plus) Compile(gen generator.Generator[int]) *automata.NFA[int] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)

	subNfa := p.Subexp.Compile(gen)
	epsilonTransitions := maps.Clone(subNfa.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[int][]int)
	}

	epsilonTransitions[intialState] = append(epsilonTransitions[intialState], subNfa.IntialState)
	for fs := range subNfa.FinalStates.Iter() {
		epsilonTransitions[fs] = append(epsilonTransitions[fs], finalState, subNfa.IntialState)
	}

	return &automata.NFA[int]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates.Union(subNfa.AllStates),
		Alphabet:           subNfa.Alphabet,
		Delta:              subNfa.Delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (p Plus) Optimize() Regex { return p }

type Cat struct {
	Left  Regex
	Right Regex
}

func (Cat) Opcode() Opcode { return CatOp }
func (c Cat) Compile(gen generator.Generator[int]) *automata.NFA[int] {
	lc := c.Left.Compile(gen)
	rc := c.Right.Compile(gen)
	allStates := mapset.NewSet[int]().Union(lc.AllStates).Union(rc.AllStates)
	alphabet := mapset.NewSet[automata.Symbol]().Union(lc.Alphabet).Union(rc.Alphabet)

	delta := maps.Clone(lc.Delta)
	maps.Insert(delta, maps.All(rc.Delta))

	epsilonTransitions := maps.Clone(lc.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[int][]int)
	}
	maps.Insert(epsilonTransitions, maps.All(rc.EpsilonTransitions))
	for fs := range lc.FinalStates.Iter() {
		if epsilonTransitions[fs] == nil {
		}
		epsilonTransitions[fs] = append(epsilonTransitions[fs], rc.IntialState)
	}

	return &automata.NFA[int]{
		IntialState:        lc.IntialState,
		FinalStates:        rc.FinalStates,
		AllStates:          allStates,
		Alphabet:           alphabet,
		Delta:              delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (c Cat) Optimize() Regex {
	return Cat{
		Left:  c.Left.Optimize(),
		Right: c.Right.Optimize(),
	}
}

type Maybe struct {
	Subexp Regex
}

func (Maybe) Opcode() Opcode { return MaybeOp }
func (m Maybe) Compile(gen generator.Generator[int]) *automata.NFA[int] {
	intialState := gen.Generate()
	finalState := gen.Generate()
	allStates := mapset.NewSet(intialState, finalState)

	subNfa := m.Subexp.Compile(gen)
	epsilonTransitions := maps.Clone(subNfa.EpsilonTransitions)
	if epsilonTransitions == nil {
		epsilonTransitions = make(map[int][]int)
	}

	epsilonTransitions[intialState] = append(epsilonTransitions[intialState], finalState, subNfa.IntialState)
	for fs := range subNfa.FinalStates.Iter() {
		epsilonTransitions[fs] = append(epsilonTransitions[fs], finalState)
	}

	return &automata.NFA[int]{
		IntialState:        intialState,
		FinalStates:        mapset.NewSet(finalState),
		AllStates:          allStates.Union(subNfa.AllStates),
		Alphabet:           subNfa.Alphabet,
		Delta:              subNfa.Delta,
		EpsilonTransitions: epsilonTransitions,
	}
}

func (m Maybe) Optimize() Regex { return m }
