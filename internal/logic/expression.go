package logic

import (
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

func NewExpression() *Expression {
	return &Expression{
		nodes: []Node{},
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithTerm(term Term) *Expression {
	return &Expression{
		nodes: []Node{{Term: term, Rel: NewSelfRelation(0)}},
	}
}

func NewExpressionWithStringExpression(expr string) *Expression {
	return &Expression{
		nodes: []Node{}, // TODO: Заменить на вызов ExpressionParser, когда он будет готов
		rep:   "",
		mod:   true,
	}
}

func NewExpressionWithNodes(nodes []Node) *Expression {
	return &Expression{
		nodes: nodes,
		rep:   "",
		mod:   true,
	}
}

func (e *Expression) inRange(idx int) bool {
	return idx < len(e.nodes)
}

func (e *Expression) updateRep() {
	if e.Empty() {
		e.rep = "empty"
		e.mod = false
		return
	}

	var builder strings.Builder

	// TODO: Закончить

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

func (e *Expression) Normalize() {}

func (e *Expression) Standardize() {}

func (e *Expression) MakeConstant() {}

func (e *Expression) Subtree(idx int) Relation {
	if e.inRange(idx) {
		return e.nodes[idx].Rel
	}
	return NewRelation()
}

func (e *Expression) CopySubtree(idx int) Expression {
	return *NewExpression()
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

func (e *Expression) Negation(idx int) {}

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
	return *NewExpression()
}

func Construct(lhs *Expression, op Operation, rhs *Expression) Expression {
	return *NewExpression()
}

func (e *Expression) Equals(rhs Expression, varIgnore bool) bool {
	return true
}
