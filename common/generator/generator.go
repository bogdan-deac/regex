package generator

import "fmt"

type Generator[T any] interface {
	Generate() T
}

type PrintableInt int

func (i PrintableInt) String() string {
	return fmt.Sprintf("%d", i)
}

func (i *PrintableInt) Generate() PrintableInt {
	toRet := *i
	*i++
	return toRet
}
func NewIntGenerator() Generator[PrintableInt] {
	// by using 1 as the default start state, we can optimize dfa construction by treating 0 as a non-transition
	var start PrintableInt = 1
	return new(start)
}
