package expression

import (
	"logical-inference/internal/pkg/queue"
	"strings"
)

type Node struct {
	Term     Term
	Relation Relation
}

type Expression struct {
	Nodes []Node
	rep   string
	mod   bool
}

func NewExpression() *Expression {
	return &Expression{
		Nodes: make([]Node, 0),
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithTerm(term Term) *Expression {
	return &Expression{
		Nodes: []Node{{Term: term, Relation: *NewSelfRelation(0)}},
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithNodes(nodes []Node) *Expression {
	nodesCopy := make([]Node, len(nodes))
	copy(nodesCopy, nodes)

	return &Expression{
		Nodes: nodesCopy,
		rep:   "",
		mod:   true,
	}
}

func (e *Expression) inRange(idx uint) bool {
	return idx < uint(len(e.Nodes))
}

func (e *Expression) updateRep() {
	if e.Empty() {
		e.rep = "empty"
		e.mod = false
		return
	}

	var builder strings.Builder

	traverse := func() func(root Relation) {
		var f func(root Relation)
		f = func(root Relation) {
			if root.Self() == invalidIdx {
				return
			}

			brackets := root.Parent() != invalidIdx && e.Nodes[root.Self()].Term.Type == Function
			if brackets {
				builder.WriteString("(")
			}

			f(e.Subtree(root.Left()))
			builder.WriteString(e.Nodes[root.Self()].Term.String())
			f(e.Subtree(root.Right()))

			if brackets {
				builder.WriteString(")")
			}

			return
		}
		return f
	}()

	traverse(e.Subtree(0))
	e.rep = builder.String()
	e.mod = false
}

func (e *Expression) Empty() bool {
	return len(e.Nodes) == 0
}

func (e *Expression) Size() int {
	return len(e.Nodes)
}

func (e *Expression) String() string {
	if e.mod {
		e.updateRep()
	}
	return e.rep
}

func (e *Expression) Operations(op Operation) uint {
	count := uint(0)
	for _, node := range e.Nodes {
		if node.Term.Type != Function {
			continue
		}

		if node.Term.Op == op {
			count++
		}
	}
	return count
}

func (e *Expression) Variables() []Value {
	result := make([]Value, 0, len(e.Nodes))
	for _, node := range e.Nodes {
		if node.Term.Type == Variable {
			result = append(result, node.Term.Val)
		}
	}
	return result
}

func (e *Expression) MaxValue() Value {
	value := Value(0)
	for _, node := range e.Nodes {
		if node.Term.Type == Variable {
			value = max(value, node.Term.Val)
		}
	}
	return value
}

func (e *Expression) MinValue() Value {
	value := Value(-1)
	for _, node := range e.Nodes {
		if node.Term.Type == Variable {
			if value == -1 {
				value = node.Term.Val
			} else {
				value = min(value, node.Term.Val)
			}
		}
	}
	return value
}

func (e *Expression) Normalize() {
	order := make([]Value, 0)
	remapping := make(map[Value]Value)

	traverse := func() func(node Relation) {
		var f func(node Relation)
		f = func(node Relation) {
			if node.Self() == invalidIdx {
				return
			}

			f(e.Subtree(node.Left()))

			if e.Nodes[node.Self()].Term.Type == Variable {
				order = append(order, e.Nodes[node.Self()].Term.Val)
			}

			f(e.Subtree(node.Right()))
		}
		return f
	}()
	traverse(e.Subtree(0))

	newVal := Value(1)
	for _, entry := range order {
		if _, ok := remapping[entry]; !ok {
			remapping[entry] = newVal
			newVal++
		}
	}

	for i := range e.Nodes {
		if e.Nodes[i].Term.Type == Variable {
			e.Nodes[i].Term.Val = remapping[e.Nodes[i].Term.Val]
		}
	}

	e.mod = true
}

func (e *Expression) Standardize() {
	q := queue.New[uint]()
	q.Push(0)

	for q.Len() > 0 {
		nodeIdx := *q.Pop()

		if nodeIdx == invalidIdx {
			continue
		}

		if e.Nodes[nodeIdx].Term.Type != Function {
			continue
		}

		if e.Nodes[nodeIdx].Term.Op == Disjunction {
			e.Nodes[nodeIdx].Term.Op = Implication
			e.Negation(e.Subtree(nodeIdx).Left())
		}

		if e.HasLeft(nodeIdx) {
			q.Push(e.Subtree(nodeIdx).Left())
		}
		if e.HasRight(nodeIdx) {
			q.Push(e.Subtree(nodeIdx).Right())
		}
	}
	e.mod = true
}

func (e *Expression) MakeConst() {
	for i := range e.Nodes {
		if e.Nodes[i].Term.Type == Variable {
			e.Nodes[i].Term.Type = Constant
		}
	}
	e.mod = true
}

func (e *Expression) Subtree(idx uint) Relation {
	if e.inRange(idx) {
		return e.Nodes[idx].Relation
	}
	return *NewRelation()
}

func (e *Expression) CopySubtree(idx uint) *Expression {
	newRootIdx := e.Subtree(idx).Self()
	nodes := make([]Node, 0)
	remapping := make(map[uint]uint)
	i := uint(0)

	traverse := func() func(node Relation) {
		var f func(node Relation)
		f = func(node Relation) {
			if node.Self() == invalidIdx {
				return
			}

			nodes = append(nodes, e.Nodes[node.Self()])
			remapping[node.Self()] = i
			i++

			f(e.Subtree(node.Left()))
			f(e.Subtree(node.Right()))
		}
		return f
	}()

	traverse(e.Subtree(newRootIdx))
	nodes[0].Relation.Refs[ParentIdx] = invalidIdx

	for j := range nodes {
		for k := range nodes[j].Relation.Refs {
			if ref, ok := remapping[nodes[j].Relation.Refs[k]]; ok {
				nodes[j].Relation.Refs[k] = ref
			}
		}
	}
	return NewExpressionWithNodes(nodes)
}

func (e *Expression) Contains(term Term) bool {
	if term.Type != Variable && term.Type != Constant {
		return false
	}

	for _, node := range e.Nodes {
		if node.Term.Type != Variable && node.Term.Type != Constant {
			continue
		}

		if node.Term.Val == term.Val {
			return true
		}
	}
	return false
}

func (e *Expression) HasLeft(idx uint) bool {
	if !e.inRange(idx) {
		return false
	}
	return e.inRange(e.Nodes[idx].Relation.Left())
}

func (e *Expression) HasRight(idx uint) bool {
	if !e.inRange(idx) {
		return false
	}
	return e.inRange(e.Nodes[idx].Relation.Right())
}

func (e *Expression) Negation(idx uint) {
	if !e.inRange(idx) {
		return
	}

	q := queue.New[uint]()
	q.Push(idx)

	for q.Len() > 0 {
		nodeIdx := *q.Pop()

		if nodeIdx == invalidIdx {
			continue
		}

		if e.Nodes[nodeIdx].Term.Type != Function {
			if e.Nodes[nodeIdx].Term.Op == Negation {
				e.Nodes[nodeIdx].Term.Op = Nop
			} else {
				e.Nodes[nodeIdx].Term.Op = Negation
			}
			continue
		}

		e.Nodes[nodeIdx].Term.Op = e.Nodes[nodeIdx].Term.Op.Opposite()

		if e.Nodes[nodeIdx].Term.Op == Implication || e.Nodes[nodeIdx].Term.Op == Conjunction {
			q.Push(e.Subtree(nodeIdx).Right())
		} else if e.Nodes[nodeIdx].Term.Op == Disjunction {
			q.Push(e.Subtree(nodeIdx).Left())
			q.Push(e.Subtree(nodeIdx).Right())
		}
	}

	e.mod = true
}

func (e *Expression) ChangeVariables(bound Value) {
	bound -= e.MinValue()
	for i := range e.Nodes {
		if e.Nodes[i].Term.Type == Variable {
			e.Nodes[i].Term.Val += bound
		}
	}
	e.mod = true
}

func (e *Expression) Copy() Expression {
	newE := Expression{
		Nodes: make([]Node, len(e.Nodes)),
		mod:   true,
		rep:   e.rep,
	}

	if e.Nodes != nil {
		copy(newE.Nodes, e.Nodes)
	}
	return newE
}

func (e *Expression) Replace(val Value, expr Expression) Expression {
	if expr.Empty() {
		return *e
	}

	indices := make([]uint, 0)
	newExpr := expr.Copy()
	newExprNeg := newExpr.Copy()
	newExprNeg.Negation(0)

	appropriateVal := Value(0)

	for _, node := range e.Nodes {
		if node.Term.Type == Variable && node.Term.Val == val {
			indices = append(indices, node.Relation.Self())
			appropriateVal = max(appropriateVal, -appropriateVal, node.Term.Val)
		}
	}

	if len(indices) == 0 {
		return *e
	}

	offset := e.Size()
	appropriateVal += 1
	for _, entry := range indices {
		replacement := newExpr
		if e.Nodes[entry].Term.Op == Negation {
			replacement = newExprNeg
		}

		e.Nodes[entry] = Node{
			Term: replacement.Nodes[0].Term,
			Relation: *NewRelationWithIndices(
				e.Nodes[entry].Relation.Self(),
				increaseIdx(replacement.Subtree(0).Left(), uint(offset)-1),
				increaseIdx(replacement.Subtree(0).Right(), uint(offset)-1),
				e.Nodes[entry].Relation.Parent(),
			),
		}

		for i := 1; i < replacement.Size(); i++ {
			e.Nodes = append(e.Nodes, replacement.Nodes[i])

			for j := range e.Nodes[len(e.Nodes)-1].Relation.Refs {
				e.Nodes[len(e.Nodes)-1].Relation.Refs[j] = increaseIdx(e.Nodes[len(e.Nodes)-1].Relation.Refs[j], uint(offset)-1)
			}
		}

		if e.Subtree(entry).Left() != invalidIdx {
			e.Nodes[e.Subtree(entry).Left()].Relation.Refs[ParentIdx] = entry
		}
		if e.Subtree(entry).Right() != invalidIdx {
			e.Nodes[e.Subtree(entry).Right()].Relation.Refs[ParentIdx] = entry
		}

		offset = len(e.Nodes)
	}

	e.mod = true
	return *e
}

func Construct(lhs Expression, op Operation, rhs Expression) Expression {
	expr := NewExpression()
	offset := uint(1)

	expr.Nodes = append(expr.Nodes, Node{
		Term:     Term{Function, op, Value(0)},
		Relation: *NewRelationWithIndices(0, 1, uint(lhs.Size())+1, invalidIdx),
	})

	processNodes := func(nodes []Node, offset uint) {
		for _, node := range nodes {
			expr.Nodes = append(expr.Nodes, node)

			for i := range expr.Nodes[len(expr.Nodes)-1].Relation.Refs {
				expr.Nodes[len(expr.Nodes)-1].Relation.Refs[i] = increaseIdx(expr.Nodes[len(expr.Nodes)-1].Relation.Refs[i], offset)
			}

			if expr.Nodes[len(expr.Nodes)-1].Relation.Parent() == invalidIdx {
				expr.Nodes[len(expr.Nodes)-1].Relation.Refs[ParentIdx] = 0
			}
		}
	}

	processNodes(lhs.Nodes, offset)
	offset += uint(lhs.Size())
	processNodes(rhs.Nodes, offset)

	expr.mod = true
	return *expr
}

func (e *Expression) Equals(other Expression, varIgnore bool) bool {
	if e.Size() != other.Size() {
		return false
	}

	for i := 0; i < e.Size(); i++ {
		if (e.Nodes[i].Term.Type == Function) != (other.Nodes[i].Term.Type == Function) {
			return false
		}

		if e.Nodes[i].Term.Type == Function && e.Nodes[i].Term.Op != other.Nodes[i].Term.Op {
			return false
		}

		if !varIgnore && (e.Nodes[i].Term.Type != other.Nodes[i].Term.Type) {
			return false
		}

		if e.Nodes[i].Term.Val != other.Nodes[i].Term.Val || e.Nodes[i].Term.Op != other.Nodes[i].Term.Op {
			return false
		}
	}
	return true
}
