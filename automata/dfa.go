package automata

import (
	"cmp"
	"fmt"

	set "github.com/deckarep/golang-set/v2"
)

type DFA[T cmp.Ordered] struct {
	InitialState T
	FinalStates  set.Set[T]
	AllStates    set.Set[T]
	Delta        map[T]map[Symbol]T
	Alphabet     set.Set[Symbol]
}

func NewDFA[T cmp.Ordered](
	InitialState T,
	FinalStates set.Set[T],
	AllStates set.Set[T],
	Delta map[T]map[Symbol]T,
) *DFA[T] {
	return &DFA[T]{
		InitialState: InitialState,
		FinalStates:  FinalStates,
		AllStates:    AllStates,
		Delta:        Delta,
	}
}

func (dfa *DFA[T]) MapStates(f func(T) T) *DFA[T] {
	newFinalStates := set.NewSet[T]()
	for state := range dfa.FinalStates.Iter() {
		newFinalStates.Add(f(state))
	}
	newAllStates := set.NewSet[T]()
	for state := range dfa.AllStates.Iter() {
		newAllStates.Add(f(state))
	}
	newDelta := make(map[T]map[Symbol]T)
	for src, dest := range dfa.Delta {
		newSrc := f(src)
		newMap := make(map[Symbol]T, len(dest))
		for sym, state := range dest {
			newMap[sym] = f(state)
		}
		newDelta[newSrc] = newMap
	}

	return &DFA[T]{
		InitialState: f(dfa.InitialState),
		FinalStates:  newFinalStates,
		AllStates:    newAllStates,
		Delta:        newDelta,
		Alphabet:     dfa.Alphabet,
	}
}

func (dfa *DFA[T]) Accepts(input []Symbol) bool {
	currentState := dfa.InitialState
	for _, symbol := range input {
		if _, ok := dfa.Delta[currentState][symbol]; !ok {
			return false
		}
		currentState = dfa.Delta[currentState][symbol]
	}
	return dfa.FinalStates.Contains(currentState)
}

// Thompson's algorithm should not generate any unreachable state, but this is general automata functionality
func (dfa *DFA[T]) RemoveUnreachableStates() *DFA[T] {
	reachableStates := set.NewSet(dfa.InitialState)
	newStates := set.NewSet(dfa.InitialState)

	for !newStates.IsEmpty() {
		temp := set.NewSet[T]()
		for state := range newStates.Iter() {
			for sym := range dfa.Alphabet.Iter() {
				temp.Add(dfa.Delta[state][sym])
			}
		}
		newStates = temp.Difference(reachableStates)
		reachableStates = reachableStates.Union(newStates)
	}
	unreachableStates := dfa.AllStates.Difference(reachableStates)

	// since unreachable states have no transition into them, we only have to update the set of initial and final states
	dfa.AllStates = reachableStates
	dfa.FinalStates = dfa.FinalStates.Difference(unreachableStates)
	return dfa
}

// Hopcroft's algorithm for DFA minimization
func (dfa *DFA[T]) Minimize() *DFA[T] {
	// the partitions are groupings of identical states from the original DFA
	partitions := []set.Set[T]{dfa.FinalStates, dfa.AllStates.Difference(dfa.FinalStates)}
	changed := true
	// wait until the number of partitions stablizes
	for changed {
		changed = false
		newPartitions := make([]set.Set[T], 0)

		// iterate through all partitions
		for _, partition := range partitions {
			subGroups := make(map[string]set.Set[T])
			// iterate through all states in the current partition
			for state := range partition.Iter() {
				// build up key for merged states
				signature := ""
				// we want to see how the state behaves for all symbols in the alphabet
				for sym := range dfa.Alphabet.Iter() {
					if nextMapping, ok := dfa.Delta[state]; ok {
						if nextState, ok := nextMapping[sym]; ok {
							for i, part := range partitions {
								// if it has a transition to a state in a different partition, mark it
								if part.Contains(nextState) {
									signature += fmt.Sprintf("%d,", i)
									break
								}
							}
						}
					} else {
						// if no transition, then mark that as well
						signature += "X,"
					}
				}
				if subGroups[signature] == nil {
					subGroups[signature] = set.NewSet[T]()
				}
				// at the end, we know the partition where that state leads for each symbol
				// For example 1 -> "a" -> 2, 1->"b"->1 with alphabet abc has the following
				// signature: 2,1,X,
				subGroups[signature].Add(state)
			}

			// all states that have transitions inside the same partitions can form their own partition
			for _, subGroup := range subGroups {
				newPartitions = append(newPartitions, subGroup)
			}
			changed = len(newPartitions) != len(partitions)
		}
		partitions = newPartitions
	}

	// create mapping based on partition groups - no need to generate new states, take a random one
	stateMap := make(map[T]T, len(partitions))
	for _, p := range partitions {
		joinState, ok := p.Pop()
		if !ok {
			continue
		}
		stateMap[joinState] = joinState
		for state := range p.Iter() {
			stateMap[state] = joinState
		}
	}
	newFinalStates := set.NewSet[T]()
	for st := range dfa.FinalStates.Iter() {
		newFinalStates.Add(stateMap[st])
	}
	newAllStates := set.NewSet[T]()
	for st := range dfa.AllStates.Iter() {
		newAllStates.Add(stateMap[st])
	}
	newDelta := make(map[T]map[Symbol]T)
	for originState, symMapping := range dfa.Delta {
		newOriginState := stateMap[originState]
		if newDelta[newOriginState] == nil {
			newDelta[newOriginState] = make(map[Symbol]T)
		}
		for sym, destinationState := range symMapping {
			newDestinationState := stateMap[destinationState]
			newDelta[newOriginState][sym] = newDestinationState
		}
	}
	return &DFA[T]{
		InitialState: stateMap[dfa.InitialState],
		FinalStates:  newFinalStates,
		AllStates:    newAllStates,
		Delta:        newDelta,
		Alphabet:     dfa.Alphabet,
	}
}
