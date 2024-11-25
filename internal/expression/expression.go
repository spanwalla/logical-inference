package expression

import (
	"container/list"
	"strings"
)

type Node struct {
	Term Term
	Rel  Relation
}

type Expression struct {
	Nodes []Node
	rep   string
	mod   bool
}

func NewExpression() Expression {
	return Expression{
		Nodes: make([]Node, 0),
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithTerm(term Term) Expression {
	return Expression{
		Nodes: []Node{{Term: term, Rel: NewSelfRelation(0)}},
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithStringExpression(expr string) Expression {
	return Expression{
		Nodes: []Node{}, // TODO: Заменить на вызов ExpressionAnalyzer, когда он будет готов
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithNodes(nodes []Node) Expression {
	return Expression{
		Nodes: nodes,
		rep:   "",
		mod:   true,
	}
}

func (e *Expression) inRange(idx int) bool {
	return idx >= 0 && idx < len(e.Nodes)
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

func (e *Expression) Operations(op Operation) int {
	var count = 0
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
	var result = make([]Value, 0, len(e.Nodes))
	for _, node := range e.Nodes {
		if node.Term.Type == Variable {
			result = append(result, node.Term.Val)
		}
	}
	return result
}

func (e *Expression) MaxValue() Value {
	var value = Value(0)
	for _, node := range e.Nodes {
		if node.Term.Type == Variable {
			value = max(value, node.Term.Val)
		}
	}
	return value
}

func (e *Expression) MinValue() Value {
	var value = Value(-1)
	for i := range len(e.Nodes) {
		if e.Nodes[i].Term.Type == Variable {
			if value == -1 {
				value = e.Nodes[i].Term.Val
			}
			value = min(value, e.Nodes[i].Term.Val)
		}
	}
	return value
}

func (e *Expression) Normalize() {
	var order []Value
	remapping := make(map[Value]Value)

	traverse := func() func(rel Relation) {
		var f func(rel Relation)
		f = func(rel Relation) {
			if rel.Self() == invalidIdx {
				return
			}

			f(e.Subtree(rel.Left()))

			if e.Nodes[rel.Self()].Term.Type == Variable {
				order = append(order, e.Nodes[rel.Self()].Term.Val)
			}

			f(e.Subtree(rel.Right()))
		}
		return f
	}()
	traverse(e.Subtree(0))

	newVal := Value(1)
	for _, entry := range order {
		if _, ok := remapping[entry]; ok {
			continue
		}
		remapping[entry] = newVal
		newVal++
	}

	for i := range len(e.Nodes) {
		if e.Nodes[i].Term.Type == Variable {
			continue
		}
		e.Nodes[i].Term.Val = remapping[e.Nodes[i].Term.Val]
	}
	e.mod = true
}

func (e *Expression) Standardize() {
	queue := list.New()
	queue.PushBack(0)

	for queue.Len() > 0 {
		el := queue.Front()
		queue.Remove(el)

		if el.Value.(int) == invalidIdx {
			continue
		}

		if e.Nodes[el.Value.(int)].Term.Type != Function {
			continue
		}

		if e.Nodes[el.Value.(int)].Term.Op == Disjunction {
			e.Nodes[el.Value.(int)].Term.Op = Implication
			e.Negation(e.Subtree(el.Value.(int)).Left())
		}

		if e.HasLeft(el.Value.(int)) {
			queue.PushBack(e.Subtree(el.Value.(int)).Left())
		}
		if e.HasRight(el.Value.(int)) {
			queue.PushBack(e.Subtree(el.Value.(int)).Right())
		}
	}
	e.mod = true
}

func (e *Expression) MakeConst() {
	for i := range len(e.Nodes) {
		if e.Nodes[i].Term.Type == Variable {
			e.Nodes[i].Term.Type = Constant
		}
	}
	e.mod = true
}

func (e *Expression) Subtree(idx int) Relation {
	if e.inRange(idx) {
		return e.Nodes[idx].Rel
	}
	return NewRelation()
}

func (e *Expression) CopySubtree(idx int) Expression {
	newRootIdx := e.Subtree(idx).Self()
	nodes := make([]Node, 0)
	remapping := make(map[int]int)
	i := 0

	traverse := func() func(rel Relation) {
		var f func(rel Relation)
		f = func(rel Relation) {
			if rel.Self() == invalidIdx {
				return
			}

			nodes = append(nodes, e.Nodes[rel.Self()])
			remapping[rel.Self()] = i
			i++

			f(e.Subtree(rel.Left()))
			f(e.Subtree(rel.Right()))
		}
		return f
	}()

	traverse(e.Subtree(newRootIdx))
	nodes[0].Rel.Refs[ParentIdx] = invalidIdx

	for j := range len(nodes) {
		for k := range len(nodes[j].Rel.Refs) {
			if ref, ok := remapping[nodes[j].Rel.Refs[k]]; ok {
				nodes[j].Rel.Refs[k] = ref
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

func (e *Expression) HasLeft(idx int) bool {
	if !e.inRange(idx) {
		return false
	}
	return e.inRange(e.Nodes[idx].Rel.Left())
}

func (e *Expression) HasRight(idx int) bool {
	if !e.inRange(idx) {
		return false
	}
	return e.inRange(e.Nodes[idx].Rel.Right())
}

func (e *Expression) Negation(idx int) {
	if !e.inRange(idx) {
		return
	}

	queue := list.New()
	queue.PushBack(idx)

	for queue.Len() > 0 {
		el := queue.Front()
		queue.Remove(el)

		if el.Value.(int) == invalidIdx {
			continue
		}

		if e.Nodes[el.Value.(int)].Term.Type != Function {
			if e.Nodes[el.Value.(int)].Term.Op == Negation {
				e.Nodes[el.Value.(int)].Term.Op = Nop
			} else {
				e.Nodes[el.Value.(int)].Term.Op = Negation
			}
			continue
		}

		e.Nodes[el.Value.(int)].Term.Op = e.Nodes[el.Value.(int)].Term.Op.Opposite()

		if e.Nodes[el.Value.(int)].Term.Op == Implication || e.Nodes[el.Value.(int)].Term.Op == Conjunction {
			queue.PushBack(e.Subtree(el.Value.(int)).Right())
		} else if e.Nodes[el.Value.(int)].Term.Op == Disjunction {
			queue.PushBack(e.Subtree(el.Value.(int)).Left())
			queue.PushBack(e.Subtree(el.Value.(int)).Right())
		}
	}

	e.mod = true
}

func (e *Expression) ChangeVariables(bound Value) {
	bound -= e.MinValue()
	for i := range len(e.Nodes) {
		if e.Nodes[i].Term.Type == Variable {
			e.Nodes[i].Term.Val += bound
		}
	}
	e.mod = true
}

func (e *Expression) Replace(val Value, expr *Expression) Expression {
	if expr.Empty() {
		return *e
	}
	indices := make([]int, 0)
	newExpr := expr
	newExprNeg := newExpr
	newExprNeg.Negation(0)

	appropriateVal := Value(0)

	for _, node := range e.Nodes {
		if node.Term.Type == Variable && node.Term.Val == val {
			indices = append(indices, node.Rel.Self())
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
			Rel: NewRelationWithIndices(e.Nodes[entry].Rel.Self(), increaseIdx(replacement.Subtree(0).Left(), offset-1),
				decreaseIdx(replacement.Subtree(0).Right(), offset-1),
				e.Nodes[entry].Rel.Parent()),
		}

		for i := 1; i < replacement.Size(); i++ {
			e.Nodes = append(e.Nodes, replacement.Nodes[i])

			for j := range len(e.Nodes[len(e.Nodes)-1].Rel.Refs) {
				e.Nodes[len(e.Nodes)-1].Rel.Refs[j] = increaseIdx(e.Nodes[len(e.Nodes)-1].Rel.Refs[j], offset-1)
			}
		}

		if e.Subtree(entry).Left() != invalidIdx {
			e.Nodes[e.Subtree(entry).Left()].Rel.Refs[ParentIdx] = entry
		}
		if e.Subtree(entry).Right() != invalidIdx {
			e.Nodes[e.Subtree(entry).Right()].Rel.Refs[ParentIdx] = entry
		}

		offset = len(e.Nodes)
	}

	e.mod = true
	return *e
}

func Construct(lhs *Expression, op Operation, rhs *Expression) Expression {
	offset := 1
	expr := NewExpression()
	expr.Nodes = append(expr.Nodes, Node{
		Term: Term{Function, op, Value(0)},
		Rel:  NewRelationWithIndices(0, 1, lhs.Size()+1, invalidIdx),
	})

	processNodes := func(nodes []Node, offset int) {
		for _, node := range nodes {
			expr.Nodes = append(expr.Nodes, node)

			for i := range len(expr.Nodes[len(expr.Nodes)-1].Rel.Refs) {
				expr.Nodes[len(expr.Nodes)-1].Rel.Refs[i] = increaseIdx(expr.Nodes[len(expr.Nodes)-1].Rel.Refs[i], offset)
			}

			if expr.Nodes[len(expr.Nodes)-1].Rel.Parent() == invalidIdx {
				expr.Nodes[len(expr.Nodes)-1].Rel.Refs[ParentIdx] = 0
			}
		}
	}

	processNodes(lhs.Nodes, offset)
	offset += lhs.Size()
	processNodes(rhs.Nodes, offset)

	expr.mod = true
	return expr
}

func (e *Expression) Equals(other *Expression, varIgnore bool) bool {
	if e.Size() != other.Size() {
		return false
	}

	for i := 1; i < e.Size(); i++ {
		if (e.Nodes[i].Term.Type == Function) != (other.Nodes[i].Term.Type == Function) {
			return false
		}

		if e.Nodes[i].Term.Type == Function && e.Nodes[i].Term.Op != other.Nodes[i].Term.Op {
			return false
		}

		if !varIgnore && e.Nodes[i].Term.Type != other.Nodes[i].Term.Type {
			return false
		}

		if e.Nodes[i].Term.Val != other.Nodes[i].Term.Val || e.Nodes[i].Term.Op != other.Nodes[i].Term.Op {
			return false
		}
	}
	return true
}
