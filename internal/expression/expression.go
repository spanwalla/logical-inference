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
	nodes []Node
	rep   string
	mod   bool
}

func NewExpression() Expression {
	return Expression{
		nodes: make([]Node, 0),
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithTerm(term Term) Expression {
	return Expression{
		nodes: []Node{{Term: term, Rel: NewSelfRelation(0)}},
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithStringExpression(expr string) Expression {
	return Expression{
		nodes: []Node{}, // TODO: Заменить на вызов ExpressionAnalyzer, когда он будет готов
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithNodes(nodes []Node) Expression {
	return Expression{
		nodes: nodes,
		rep:   "",
		mod:   true,
	}
}

func (e *Expression) inRange(idx int) bool {
	return idx >= 0 && idx < len(e.nodes)
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

			brackets := root.Parent() != invalidIdx && e.nodes[root.Self()].Term.Type == Function
			if brackets {
				builder.WriteString("(")
			}

			f(e.Subtree(root.Left()))
			builder.WriteString(e.nodes[root.Self()].Term.String())
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

func (e *Expression) Operations(op Operation) int {
	var count = 0
	for _, node := range e.nodes {
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
	var result = make([]Value, 0, len(e.nodes))
	for _, node := range e.nodes {
		if node.Term.Type == Variable {
			result = append(result, node.Term.Val)
		}
	}
	return result
}

func (e *Expression) MaxValue() Value {
	var value = Value(0)
	for _, node := range e.nodes {
		if node.Term.Type == Variable {
			value = max(value, node.Term.Val)
		}
	}
	return value
}

func (e *Expression) MinValue() Value {
	var value = Value(-1)
	for _, node := range e.nodes {
		if node.Term.Type == Variable {
			if value == -1 {
				value = node.Term.Val
			}
			value = min(value, node.Term.Val)
		}
	}
	return value
}

func (e *Expression) Normalize() {
	var order []Value
	var remapping map[Value]Value

	traverse := func() func(rel Relation) {
		var f func(rel Relation)
		f = func(rel Relation) {
			if rel.Self() == invalidIdx {
				return
			}

			f(e.Subtree(rel.Left()))

			if e.nodes[rel.Self()].Term.Type == Variable {
				order = append(order, e.nodes[rel.Self()].Term.Val)
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

	for _, node := range e.nodes {
		if node.Term.Type == Variable {
			continue
		}
		node.Term.Val = remapping[node.Term.Val]
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

		if e.nodes[el.Value.(int)].Term.Type != Function {
			continue
		}

		if e.nodes[el.Value.(int)].Term.Op == Disjunction {
			e.nodes[el.Value.(int)].Term.Op = Implication
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
	for _, node := range e.nodes {
		if node.Term.Type == Variable {
			node.Term.Type = Constant
		}
	}
	e.mod = true
}

func (e *Expression) Subtree(idx int) Relation {
	if e.inRange(idx) {
		return e.nodes[idx].Rel
	}
	return NewRelation()
}

func (e *Expression) CopySubtree(idx int) Expression {
	newRootIdx := e.Subtree(idx).Self()
	nodes := make([]Node, 0)
	var remapping map[int]int
	i := 0

	traverse := func() func(rel Relation) {
		var f func(rel Relation)
		f = func(rel Relation) {
			if rel.Self() == invalidIdx {
				return
			}

			nodes = append(nodes, e.nodes[rel.Self()])
			remapping[rel.Self()] = i
			i++

			f(e.Subtree(rel.Left()))
			f(e.Subtree(rel.Right()))
		}
		return f
	}()

	traverse(e.Subtree(newRootIdx))
	e.nodes[0].Rel.Refs[ParentIdx] = invalidIdx

	for _, node := range nodes {
		for _, ref := range node.Rel.Refs {
			if _, ok := remapping[ref]; !ok {
				continue
			}
			ref = remapping[ref]
		}
	}
	return NewExpressionWithNodes(nodes)
}

func (e *Expression) Contains(term Term) bool {
	if term.Type != Variable && term.Type != Constant {
		return false
	}

	for _, node := range e.nodes {
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
	return e.inRange(e.nodes[idx].Rel.Left())
}

func (e *Expression) HasRight(idx int) bool {
	if !e.inRange(idx) {
		return false
	}
	return e.inRange(e.nodes[idx].Rel.Right())
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

		if e.nodes[el.Value.(int)].Term.Type != Function {
			if e.nodes[el.Value.(int)].Term.Op == Negation {
				e.nodes[el.Value.(int)].Term.Op = Nop
			} else {
				e.nodes[el.Value.(int)].Term.Op = Negation
			}
			continue
		}

		e.nodes[el.Value.(int)].Term.Op = e.nodes[el.Value.(int)].Term.Op.Opposite()

		if e.nodes[el.Value.(int)].Term.Op == Implication || e.nodes[el.Value.(int)].Term.Op == Conjunction {
			queue.PushBack(e.Subtree(el.Value.(int)).Right())
		} else if e.nodes[el.Value.(int)].Term.Op == Disjunction {
			queue.PushBack(e.Subtree(el.Value.(int)).Left())
			queue.PushBack(e.Subtree(el.Value.(int)).Right())
		}
	}

	e.mod = true
}

func (e *Expression) ChangeVariables(bound Value) {
	bound -= e.MinValue()
	for _, node := range e.nodes {
		if node.Term.Type == Variable {
			node.Term.Val += bound
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

	for _, node := range e.nodes {
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
		if e.nodes[entry].Term.Op == Negation {
			replacement = newExprNeg
		}

		e.nodes[entry] = Node{
			Term: replacement.nodes[0].Term,
			Rel: NewRelationWithIndices(e.nodes[entry].Rel.Self(), increaseIdx(replacement.Subtree(0).Left(), offset-1),
				decreaseIdx(replacement.Subtree(0).Right(), offset-1),
				e.nodes[entry].Rel.Parent()),
		}

		for i := 1; i < replacement.Size(); i++ {
			e.nodes = append(e.nodes, replacement.nodes[i])

			for j := range len(e.nodes[len(e.nodes)-1].Rel.Refs) {
				e.nodes[len(e.nodes)-1].Rel.Refs[j] = increaseIdx(e.nodes[len(e.nodes)-1].Rel.Refs[j], offset-1)
			}
		}

		if e.Subtree(entry).Left() != invalidIdx {
			e.nodes[e.Subtree(entry).Left()].Rel.Refs[ParentIdx] = entry
		}
		if e.Subtree(entry).Right() != invalidIdx {
			e.nodes[e.Subtree(entry).Right()].Rel.Refs[ParentIdx] = entry
		}

		offset = len(e.nodes)
	}

	e.mod = true
	return *e
}

func Construct(lhs *Expression, op Operation, rhs *Expression) Expression {
	offset := 1
	expr := NewExpression()
	expr.nodes = append(expr.nodes, Node{
		Term: Term{Function, op, Value(0)},
		Rel:  NewRelationWithIndices(0, 1, lhs.Size()+1, invalidIdx),
	})

	processNodes := func(nodes []Node, offset int) {
		for _, node := range nodes {
			expr.nodes = append(expr.nodes, node)

			for i := range len(expr.nodes[len(expr.nodes)-1].Rel.Refs) {
				expr.nodes[len(expr.nodes)-1].Rel.Refs[i] = increaseIdx(expr.nodes[len(expr.nodes)-1].Rel.Refs[i], offset)
			}

			if expr.nodes[len(expr.nodes)-1].Rel.Parent() == invalidIdx {
				expr.nodes[len(expr.nodes)-1].Rel.Refs[ParentIdx] = 0
			}
		}
	}

	processNodes(lhs.nodes, offset)
	offset += lhs.Size()
	processNodes(rhs.nodes, offset)

	expr.mod = true
	return expr
}

func (e *Expression) Equals(other *Expression, varIgnore bool) bool {
	if e.Size() != other.Size() {
		return false
	}

	for i := 1; i < e.Size(); i++ {
		if (e.nodes[i].Term.Type == Function) != (other.nodes[i].Term.Type == Function) {
			return false
		}

		if e.nodes[i].Term.Type == Function && e.nodes[i].Term.Op != other.nodes[i].Term.Op {
			return false
		}

		if !varIgnore && e.nodes[i].Term.Type != other.nodes[i].Term.Type {
			return false
		}

		if e.nodes[i].Term.Val != other.nodes[i].Term.Val || e.nodes[i].Term.Op != other.nodes[i].Term.Op {
			return false
		}
	}
	return true
}
