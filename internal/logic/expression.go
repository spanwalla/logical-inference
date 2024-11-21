package logic

type Node struct {
	Term Term
	Rel  Relation
}

type Expression struct {
	nodes []Node
	rep   string
	mod   bool
}

func (e *Expression) inRange(idx int) bool {
	return idx < len(e.nodes)
}

func (e *Expression) updateRep() {

}

func (e *Expression) Empty() bool {
	return len(e.nodes) == 0
}

func (e *Expression) Size() int {
	return len(e.nodes)
}

func (e *Expression) String() string {
	if e.mod {
		e.updateRep()
	}
	return e.rep
}
