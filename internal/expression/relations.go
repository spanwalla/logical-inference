package expression

const (
	SelfIdx = iota
	LeftIdx
	RightIdx
	ParentIdx
)

type Relation [4]uint // Индексы самого себя, левого, правого, родителя

func (r Relation) Self() uint {
	return r[SelfIdx]
}

func (r Relation) Left() uint {
	return r[LeftIdx]
}

func (r Relation) Right() uint {
	return r[RightIdx]
}

func (r Relation) Parent() uint {
	return r[ParentIdx]
}

func NewRelation() *Relation {
	return &Relation{invalidIdx, invalidIdx, invalidIdx, invalidIdx}
}

func NewSelfRelation(idx uint) *Relation {
	return &Relation{idx, invalidIdx, invalidIdx, invalidIdx}
}
