package helper

import (
	"container/list"
	"logical-inference/internal/expression"
)

func TopologicalSortUtil(v expression.Value, adj [][]expression.Value, visited []bool, stack *list.List) {
	visited[v] = true

	// Рекурсивный вызов для всех смежных узлов
	for _, i := range adj[v] {
		if !visited[i] {
			TopologicalSortUtil(i, adj, visited, stack)
		}
	}

	// Добавляем вершину в стек
	stack.PushFront(v)
}

// Основная функция топологической сортировки
func TopologicalSort(adj [][]expression.Value, size expression.Value) []expression.Value {
	// stack := []expression.Value{}
	stack := list.New()
	visited := make([]bool, size)
	order := make([]expression.Value, 0, size)

	// Проходим по всем узлам графа
	for i := expression.Value(0); i < size; i++ {
		if !visited[i] {
			TopologicalSortUtil(i, adj, visited, stack)
		}
	}

	for stack.Len() > 0 {
		order = append(order, stack.Front().Value.(expression.Value))
	}

	return order
}

func AddConstraint(term expression.Term, substitution expression.Expression, sub map[expression.Value]expression.Expression) bool {
	if substitution.Nodes[0].Term.Type == expression.Function && substitution.Contains(term) {
		return false
	}

	sub[term.Val] = substitution
	return true
}

func GetUnification(left, right expression.Expression, substitution map[expression.Value]expression.Expression) bool {
	var sub map[expression.Value]expression.Expression

	right.ChangeVariables(left.MaxValue() + 1)
	v := right.MaxValue() + 1

	exprQueue := list.New()
	expr := [2]int{left.Subtree(0).Self(), left.Subtree(0).Self()}
	exprQueue.PushBack(expr)

	var lhs, rhs expression.Expression

	for exprQueue.Len() > 0 {
		el := exprQueue.Front()
		exprQueue.Remove(el)

		leftIdx := el.Value.([2]int)[0]
		rightIdx := el.Value.([2]int)[1]
		leftTerm := left.Nodes[leftIdx].Term
		rightTerm := right.Nodes[rightIdx].Term

		// case 0
		if leftTerm.Type == expression.Function && rightTerm.Type == expression.Function {
			if leftTerm.Op != rightTerm.Op {
				return false
			}

			exprQueue.PushBack([2]int{left.Subtree(leftIdx).Right(), right.Subtree(rightIdx).Right()})
			exprQueue.PushBack([2]int{left.Subtree(leftIdx).Right(), right.Subtree(rightIdx).Right()})
			continue
		}

		lhs = left.CopySubtree(leftIdx)
		rhs = right.CopySubtree(rightIdx)

		for lhs.Nodes[0].Term.Type == expression.Variable {
			shouldNegate := lhs.Nodes[0].Term.Op == expression.Negation
			var ok bool
			if lhs, ok = sub[lhs.Nodes[0].Term.Val]; !ok {
				break
			}
			if shouldNegate {
				lhs.Negation(0)
			}
		}

		for rhs.Nodes[0].Term.Type == expression.Variable {
			shouldNegate := rhs.Nodes[0].Term.Op == expression.Negation
			var ok bool
			if rhs, ok = sub[rhs.Nodes[0].Term.Val]; !ok {
				break
			}
			if shouldNegate {
				rhs.Negation(0)
			}
		}

		// case 1
		if lhs.Nodes[0].Term.Type == expression.Constant && rhs.Nodes[0].Term.Type == expression.Constant {
			if lhs.Nodes[0].Term != rhs.Nodes[0].Term {
				return false
			}

			continue
		}

		// case 2
		if lhs.Nodes[0].Term.Type == expression.Constant && rhs.Nodes[0].Term.Type == expression.Variable {
			if rhs.Nodes[0].Term.Op == expression.Negation {
				rhs.Nodes[0].Term.Op = expression.Nop
				if rhs.Nodes[0].Term.Op != expression.Negation {
					rhs.Nodes[0].Term.Op = expression.Negation
				}
			}

			if !AddConstraint(lhs.Nodes[0].Term, rhs, sub) {
				return false
			}

			continue
		}

		// case 3

		// case 4

		// add new variable
		op := expression.Nop
		if lhs.Nodes[0].Term.Op == expression.Negation || rhs.Nodes[0].Term.Op == expression.Negation {
			op = expression.Negation
		}

		term := expression.Term{
			Type: expression.Variable,
			Op:   op,
			Val:  v + 1,
		}
		v++
		expr := expression.NewExpressionWithTerm(term)

		// case 5

		// case 6

		return false
	}

	// ...
	substitution = sub
	return true
}

func IsEqual(left, right expression.Expression) bool {
	if left.Size() != right.Size() {
		return false
	}

	if left.Nodes[0].Term.Op != right.Nodes[0].Term.Op {
		return false
	}

	left.Normalize()
	right.Normalize()

	return left.Equals(&right, true)
}
