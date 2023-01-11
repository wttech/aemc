package langx

type Stack[T any] struct {
	values []T
}

func NewStackWithValue[T any](initialValue T) Stack[T] {
	return Stack[T]{values: []T{initialValue}}
}

func NewStackWithValues[T any](values []T) Stack[T] {
	return Stack[T]{values: values}
}

func EmptyStack[T any]() Stack[T] {
	return Stack[T]{values: []T{}}
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.values) == 0
}

func (s *Stack[T]) Push(value T) {
	s.values = append(s.values, value)
}

func (s *Stack[T]) Pop() T {
	if s.IsEmpty() {
		panic("cannot pop value from an empty stack")
	}
	top := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return top
}
