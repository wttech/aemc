package langx

// Iterator is a consequence of missing idiom, see: <https://github.com/golang/go/discussions/56413>
type Iterator[T any] interface {
	Next() (T, bool, error)
}

func IteratorToSlice[T any](it Iterator[T]) ([]T, error) {
	var res []T
	for {
		v, ok, err := it.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		res = append(res, v)
	}
	return res, nil
}

type Stack[T any] struct {
	values []T
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.values) == 0
}

func (s *Stack[T]) Push(value T) {
	s.values = append(s.values, value)
}

func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	top := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return top, true
}
