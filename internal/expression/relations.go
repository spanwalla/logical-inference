package expression

const (
	SelfIdx = iota
	LeftIdx
	RightIdx
	ParentIdx
)

type Relation struct {
	Refs [4]uint // Индексы самого себя, левого, правого, родителя
}

func (r Relation) Self() uint {
	return r.Refs[SelfIdx]
}

func (r Relation) Left() uint {
	return r.Refs[LeftIdx]
}

func (r Relation) Right() uint {
	return r.Refs[RightIdx]
}

func (r Relation) Parent() uint {
	return r.Refs[ParentIdx]
}

func NewRelation() *Relation {
	return &Relation{
		Refs: [4]uint{invalidIdx, invalidIdx, invalidIdx, invalidIdx},
	}
}

func NewSelfRelation(idx uint) *Relation {
	return &Relation{
		Refs: [4]uint{idx, invalidIdx, invalidIdx, invalidIdx},
	}
}

func NewRelationWithIndices(self, left, right, parent uint) *Relation {
	return &Relation{
		Refs: [4]uint{self, left, right, parent},
	}
}
