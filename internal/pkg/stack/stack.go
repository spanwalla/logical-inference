package stack

import "container/list"

type Stack[T any] struct {
	data *list.List
}

func New[T any]() *Stack[T] {
	return &Stack[T]{
		data: list.New(),
	}
}

func (s *Stack[T]) Push(value T) {
	s.data.PushBack(value)
}

func (s *Stack[T]) Pop() *T {
	if s.data.Len() == 0 {
		return nil
	}
	element := s.data.Back()
	s.data.Remove(element)
	value := element.Value.(T)
	return &value
}

func (s *Stack[T]) Peek() *T {
	if s.data.Len() == 0 {
		return nil
	}
	element := s.data.Back().Value.(T)
	return &element
}

func (s *Stack[T]) Empty() bool {
	return s.data.Len() == 0
}

func (s *Stack[T]) Len() int {
	return s.data.Len()
}
