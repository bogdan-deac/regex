package generator

type Generator[T any] interface {
	Generate() T
}

type intGenerator int

func (i *intGenerator) Generate() int {
	toRet := int(*i)
	*i++
	return toRet
}
func NewIntGenerator() Generator[int] {
	return new(intGenerator)
}
