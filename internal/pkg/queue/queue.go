package queue

import "container/list"

type Queue[T any] struct {
	data *list.List
}

func New[T any]() *Queue[T] {
	return &Queue[T]{
		data: list.New(),
	}
}

func (s *Queue[T]) Push(value T) {
	s.data.PushBack(value)
}

func (s *Queue[T]) Pop() *T {
	if s.data.Len() == 0 {
		return nil
	}
	element := s.data.Front()
	s.data.Remove(element)
	value := element.Value.(T)
	return &value
}

func (s *Queue[T]) Peek() *T {
	if s.data.Len() == 0 {
		return nil
	}
	element := s.data.Front().Value.(T)
	return &element
}

func (s *Queue[T]) Empty() bool {
	return s.data.Len() == 0
}

func (s *Queue[T]) Len() int {
	return s.data.Len()
}
