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
	return new(PrintableInt)
}
