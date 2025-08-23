package automata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bogdan-deac/regex/common/generator"
	set "github.com/deckarep/golang-set/v2"
	queue "github.com/oleiade/lane/v2"
)

type NFA[T StateLike] struct {
	IntialState        T
	FinalStates        set.Set[T]
	AllStates          set.Set[T]
	Alphabet           set.Set[Symbol]
	Delta              map[T]map[Symbol][]T
	EpsilonTransitions map[T][]T
}

func NewNFA[T StateLike](
	IntialState T,
	FinalStates set.Set[T],
	AllStates set.Set[T],
	Alphabet set.Set[Symbol],
	Delta map[T]map[Symbol][]T,
	EpsilonTransitions map[T][]T,
) *NFA[T] {
	return &NFA[T]{
		IntialState:        IntialState,
		FinalStates:        FinalStates,
		AllStates:          AllStates,
		Alphabet:           Alphabet,
		Delta:              Delta,
		EpsilonTransitions: EpsilonTransitions,
	}
}

// build epsilon closures for each state. Each epsilon closure contains the originating state
func (nfa *NFA[T]) EpsilonClosures() map[T]set.Set[T] {
	epsilonClosures := make(map[T]set.Set[T], len(nfa.EpsilonTransitions))
	for _, state := range nfa.AllStates.ToSlice() {
		epsilonClosures[state] = set.NewSet(state)
	}
	for state, epsStates := range nfa.EpsilonTransitions {
		toCheck := queue.NewQueue(epsStates...)

		visitedNodes := make(map[T]struct{})
		epsilonClosures[state].Append(epsStates...)

		for toCheck.Size() > 0 {
			next, _ := toCheck.Dequeue()
			visitedNodes[next] = struct{}{}
			// iteratively build epsilon closure for all states that are reachable via epsilon transitions
			for _, t := range nfa.EpsilonTransitions[next] {
				if _, visited := visitedNodes[t]; !visited {
					epsilonClosures[state].Add(t)
					toCheck.Enqueue(t)
				}
			}

		}
	}
	return epsilonClosures
}

func (nfa *NFA[T]) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[Alphabet] %v\n", nfa.Alphabet.ToSlice()))
	sb.WriteString("[Initial State] " + nfa.IntialState.String() + "\n")
	sb.WriteString(fmt.Sprintf("[Final States] %v", nfa.FinalStates.ToSlice()) + "\n")
	sb.WriteString(fmt.Sprintf("[ALL States] %v", nfa.AllStates.ToSlice()) + "\n")
	sb.WriteString("[DELTA]\n")
	for origin, mapping := range nfa.Delta {
		for sym, dest := range mapping {
			sb.WriteString(fmt.Sprintf("%s -> %d -> %s\n", origin.String(), sym, dest))
		}
	}
	sb.WriteString("[EPS_TRANSITIONS]\n")
	for start, end := range nfa.EpsilonTransitions {
		sb.WriteString(fmt.Sprintf("%v -> %v", start, end) + "\n")
	}
	return sb.String()
}

func (nfa *NFA[T]) RemoveWildcards() {
	nfa.Alphabet.Remove(Wildcard)
	hasWildcard := false
	for start, mapping := range nfa.Delta {
		for sym, dest := range mapping {
			if sym != Wildcard {
				continue
			}
			hasWildcard = true
			for _, alphabetSym := range ASCIIChars {
				for _, newSt := range dest {
					if !slices.Contains(mapping[alphabetSym], newSt) {
						mapping[alphabetSym] = append(mapping[alphabetSym], newSt)
					}
				}
				delete(mapping, sym)
			}
			nfa.Delta[start] = mapping
		}
	}
	if hasWildcard {
		nfa.Alphabet.Append(ASCIIChars...)
	}
}

// implemented using the subset construction algorithm
func (nfa *NFA[T]) ToDFA(g generator.Generator[T]) *DFA[T] {

	epsClosures := nfa.EpsilonClosures()
	nfa.RemoveWildcards()
	// use a trie for generating DFA states for sets of NFA states
	mergeStates := make(map[string]T)

	dfaAllStates := set.NewSet[T]()
	dfaFinalStates := set.NewSet[T]()

	dfaDelta := make(map[T]map[Symbol]T)

	initialStateWithClosure := epsClosures[nfa.IntialState]

	sliceISWC := initialStateWithClosure.ToSlice()
	slices.Sort(sliceISWC)

	dfaInitialState := g.Generate()
	mergeStates[fmt.Sprint(sliceISWC)] = dfaInitialState

	dfaAllStates.Add(dfaInitialState)

	// if any state in the epsilon-closed set, add the newly generated state to the final states as well
	if nfa.FinalStates.ContainsAny(sliceISWC...) {
		dfaFinalStates.Add(dfaInitialState)
	}

	// var leadsToSink bool
	var mergedStateValue T

	// use queue for keeping track of subsets of states
	toProcess := queue.NewQueue(sliceISWC)

	for toProcess.Size() > 0 {
		currentStateSlice, _ := toProcess.Dequeue()
		slices.Sort(currentStateSlice)
		var originState T
		if state, ok := mergeStates[fmt.Sprint(currentStateSlice)]; ok {
			originState = state
		} else {
			originState = g.Generate()
			mergeStates[fmt.Sprint(currentStateSlice)] = originState
		}
		if dfaDelta[originState] == nil {
			dfaDelta[originState] = make(map[Symbol]T)
		}
		// For each symbol, for each state, we need to analyze all paths and build states accordingly
		for symbol := range nfa.Alphabet.Iter() {
			// build eps closures for all transitions
			allTransitionsWithEps := set.NewSet[T]()
			for _, state := range currentStateSlice {
				transitions, okT := nfa.Delta[state]
				if !okT {
					continue
				}

				symTransitions, okTS := transitions[symbol]
				if !okTS {
					continue
				}

				for _, st := range symTransitions {
					allTransitionsWithEps.Append(epsClosures[st].ToSlice()...)
				}

				stateSlice := allTransitionsWithEps.ToSlice()
				slices.Sort(stateSlice)

				// if the set of states has already been processed - don't requeue it
				var ok bool
				if mergedStateValue, ok = mergeStates[fmt.Sprint(stateSlice)]; !ok {
					mergedStateValue = g.Generate()
					mergeStates[fmt.Sprint(stateSlice)] = mergedStateValue
					toProcess.Enqueue(stateSlice)
				}

				// add newly generated state to all states
				dfaAllStates.Add(mergedStateValue)

				// add to final states if the set contains any final state
				if nfa.FinalStates.ContainsAny(stateSlice...) {
					dfaFinalStates.Add(mergedStateValue)
				}

				if _, ok := dfaDelta[mergedStateValue]; !ok {
					dfaDelta[mergedStateValue] = make(map[Symbol]T)
				}
				// create transition from origin to newly generated state
				dfaDelta[originState][symbol] = mergedStateValue
			}
		}
	}

	dfa := &DFA[T]{
		InitialState: dfaInitialState,
		FinalStates:  dfaFinalStates,
		AllStates:    dfaAllStates,
		Delta:        dfaDelta,
		Alphabet:     nfa.Alphabet,
	}

	return dfa
}
