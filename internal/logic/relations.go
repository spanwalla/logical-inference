package logic

import "fmt"

const (
	SelfIdx = iota
	LeftIdx
	RightIdx
	ParentIdx
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
	return r.getRef(SelfIdx)
}

func (r Relation) Left() int {
	return r.getRef(LeftIdx)
}

func (r Relation) Right() int {
	return r.getRef(RightIdx)
}

func (r Relation) Parent() int {
	return r.getRef(ParentIdx)
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

func NewRelationWithIndices(self, left, right, parent int) Relation {
	return Relation{
		Refs: [4]int{self, left, right, parent},
	}
}
