package helper

import (
	"logical-inference/internal/expression"
	"logical-inference/internal/pkg/queue"
	"logical-inference/internal/pkg/stack"
)

func topologicalSortUtil(v expression.Value, adj [][]expression.Value, visited []bool, s *stack.Stack[expression.Value]) {
	visited[v] = true

	// Рекурсивный вызов для всех смежных узлов
	for _, i := range adj[v] {
		if !visited[i] {
			topologicalSortUtil(i, adj, visited, s)
		}
	}

	// Добавляем вершину в стек
	s.Push(v)
}

// TopologicalSort Основная функция топологической сортировки
func TopologicalSort(adj [][]expression.Value, size expression.Value) []expression.Value {
	s := stack.New[expression.Value]()
	visited := make([]bool, size)
	order := make([]expression.Value, 0, size)

	// Проходим по всем узлам графа
	for i := expression.Value(0); i < size; i++ {
		if !visited[i] {
			topologicalSortUtil(i, adj, visited, s)
		}
	}

	for s.Len() > 0 {
		el := *s.Pop()
		order = append(order, el)
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

func GetUnification(left, right expression.Expression, substitution *map[expression.Value]expression.Expression) bool {
	sub := make(map[expression.Value]expression.Expression)

	right.ChangeVariables(left.MaxValue() + 1)
	v := right.MaxValue() + 1

	exprQueue := queue.New[[2]uint]()
	exprQueue.Push([2]uint{left.Subtree(0).Self(), right.Subtree(0).Self()})

	var lhs, rhs expression.Expression

	for exprQueue.Len() > 0 {
		el := *exprQueue.Pop()

		leftIdx, rightIdx := el[0], el[1]
		leftTerm, rightTerm := left.Nodes[leftIdx].Term, right.Nodes[rightIdx].Term

		// case 0
		if leftTerm.Type == expression.Function && rightTerm.Type == expression.Function {
			if leftTerm.Op != rightTerm.Op {
				return false
			}

			exprQueue.Push([2]uint{left.Subtree(leftIdx).Left(), right.Subtree(rightIdx).Left()})
			exprQueue.Push([2]uint{left.Subtree(leftIdx).Right(), right.Subtree(rightIdx).Right()})

			continue
		}

		lhs = *left.CopySubtree(leftIdx)
		rhs = *right.CopySubtree(rightIdx)

		contains := func(key expression.Value) bool {
			_, ok := sub[key]
			return ok
		}

		for lhs.Nodes[0].Term.Type == expression.Variable && contains(lhs.Nodes[0].Term.Val) {
			shouldNegate := lhs.Nodes[0].Term.Op == expression.Negation
			lhs = sub[lhs.Nodes[0].Term.Val]
			if shouldNegate {
				lhs.Negation(0)
			}
		}

		for rhs.Nodes[0].Term.Type == expression.Variable && contains(rhs.Nodes[0].Term.Val) {
			shouldNegate := rhs.Nodes[0].Term.Op == expression.Negation
			rhs = sub[rhs.Nodes[0].Term.Val]
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
				lhs.Nodes[0].Term.Op = expression.Nop
				if lhs.Nodes[0].Term.Op != expression.Negation {
					lhs.Nodes[0].Term.Op = expression.Negation
				}
			}

			if !AddConstraint(rhs.Nodes[0].Term, lhs, sub) {
				return false
			}

			continue
		}

		// case 3
		if lhs.Nodes[0].Term.Type == expression.Variable && rhs.Nodes[0].Term.Type == expression.Constant {
			if lhs.Nodes[0].Term.Op == expression.Negation {
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

		// case 4
		if lhs.Nodes[0].Term.Type == expression.Variable && rhs.Nodes[0].Term.Type == expression.Variable {
			if lhs.Nodes[0].Term.Val == rhs.Nodes[0].Term.Val {
				if lhs.Nodes[0].Term.Op != rhs.Nodes[0].Term.Op {
					return false
				}

				continue
			}

			// add new variable
			op := expression.Nop
			if lhs.Nodes[0].Term.Op == expression.Negation || rhs.Nodes[0].Term.Op == expression.Negation {
				op = expression.Negation
			}

			term := expression.Term{
				Type: expression.Variable,
				Op:   op,
				Val:  v,
			}
			v++
			expr := *expression.NewExpressionWithTerm(term)
			negExpr := expr
			negExpr.Negation(0)

			if lhs.Nodes[0].Term.Op == expression.Negation {
				AddConstraint(lhs.Nodes[0].Term, negExpr, sub)
			} else {
				AddConstraint(lhs.Nodes[0].Term, expr, sub)
			}
			if rhs.Nodes[0].Term.Op == expression.Negation {
				AddConstraint(rhs.Nodes[0].Term, negExpr, sub)
			} else {
				AddConstraint(rhs.Nodes[0].Term, expr, sub)
			}

			continue
		}

		// case 5
		if lhs.Nodes[0].Term.Type == expression.Function {
			if rhs.Nodes[0].Term.Type != expression.Variable {
				return false
			}

			if rhs.Nodes[0].Term.Op == expression.Negation {
				lhs.Negation(0)
			}

			if !AddConstraint(rhs.Nodes[0].Term, lhs, sub) {
				return false
			}

			continue
		}

		// case 6
		if rhs.Nodes[0].Term.Type == expression.Function {
			if lhs.Nodes[0].Term.Type != expression.Variable {
				return false
			}

			if lhs.Nodes[0].Term.Op == expression.Negation {
				rhs.Negation(0)
			}

			if !AddConstraint(lhs.Nodes[0].Term, rhs, sub) {
				return false
			}

			continue
		}

		return false
	}

	adjacent := make([][]expression.Value, v-1)
	for u, expr := range sub {
		for _, w := range expr.Variables() {
			adjacent[w-1] = append(adjacent[w-1], u-1)
		}
	}

	order := TopologicalSort(adjacent, v-1)
	for i := range order {
		order[i] = order[i] + 1

		if _, exists := sub[order[i]]; !exists {
			continue
		}

		expr := sub[order[i]]
		if expr.Nodes[0].Term.Type != expression.Function {
			continue
		}

		for _, value := range expr.Variables() {
			if _, exist := sub[value]; !exist {
				continue
			}
			replacement := sub[value]
			for replacement.Nodes[0].Term.Type == expression.Variable {
				if _, exists := sub[replacement.Nodes[0].Term.Val]; !exists {
					break
				}
				shouldNegate := replacement.Nodes[0].Term.Op == expression.Negation
				replacement = sub[replacement.Nodes[0].Term.Val]
				if shouldNegate {
					replacement.Negation(0)
				}
			}

			toCheck := expression.Term{
				Type: expression.Variable,
				Op:   expression.Nop,
				Val:  value,
			}
			if replacement.Contains(toCheck) {
				return false
			}

			expr.Replace(value, replacement)
		}
	}

	*substitution = sub
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

	return left.Equals(right, true)
}
