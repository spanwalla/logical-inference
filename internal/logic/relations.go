package logic

import "fmt"

const (
	SelfIndex = iota
	LeftIndex
	RightIndex
	ParentIndex
)

type Relation struct {
	Refs [4]int // Индексы самого себя, левого, правого, родителя
}

func (r Relation) getRef(index int) int {
	if index < 0 || index >= len(r.Refs) {
		panic(fmt.Sprintf("index %d out of range", index))
	}
	return r.Refs[index]
}

func (r Relation) Self() int {
	return r.getRef(SelfIndex)
}

func (r Relation) Left() int {
	return r.getRef(LeftIndex)
}

func (r Relation) Right() int {
	return r.getRef(RightIndex)
}

func (r Relation) Parent() int {
	return r.getRef(ParentIndex)
}

func NewRelation() Relation {
	return Relation{
		Refs: [4]int{invalidIdx, invalidIdx, invalidIdx, invalidIdx},
	}
}

func NewSelfRelation(idx int) Relation {
	return Relation{
		Refs: [4]int{idx, invalidIdx, invalidIdx, invalidIdx},
	}
}
