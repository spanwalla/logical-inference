package helper

import (
	"container/list"
	"logical-inference/internal/expression"
)

func topologicalSortUtil(v expression.Value, adj [][]expression.Value, visited []bool, stack *list.List) {
	visited[v] = true

	// Рекурсивный вызов для всех смежных узлов
	for _, i := range adj[v] {
		if !visited[i] {
			topologicalSortUtil(i, adj, visited, stack)
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
			topologicalSortUtil(i, adj, visited, stack)
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

func GetUnification(left, right expression.Expression, substitution *map[expression.Value]expression.Expression) bool {
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
			expr := expression.NewExpressionWithTerm(term)
			neg_expr := expr
			neg_expr.Negation(0)

			if lhs.Nodes[0].Term.Op == expression.Negation {
				AddConstraint(lhs.Nodes[0].Term, neg_expr, sub)
			} else {
				AddConstraint(lhs.Nodes[0].Term, expr, sub)
			}
			if rhs.Nodes[0].Term.Op == expression.Negation {
				AddConstraint(rhs.Nodes[0].Term, neg_expr, sub)
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
		for w := range expr.Variables() {
			adjacent[w-1] = append(adjacent[w-1], u-1)
		}
	}

	order := TopologicalSort(adjacent, v-1)
	for _, variable := range order {
		variable := variable + 1

		if _, exists := sub[variable]; !exists {
			continue
		}

		expr, _ := sub[variable]
		if expr.Nodes[0].Term.Type != expression.Function {
			continue
		}

		for _, v := range expr.Variables() {
			if _, exist := sub[v]; !exist {
				continue
			}
			replacement, _ := sub[v]
			for replacement.Nodes[0].Term.Type == expression.Variable {
				if _, exists := sub[replacement.Nodes[0].Term.Val]; !exists {
					break
				}
				should_negate := replacement.Nodes[0].Term.Op == expression.Negation
				replacement, _ := sub[replacement.Nodes[0].Term.Val]
				if should_negate {
					replacement.Negation(0)
				}
			}
			toCheck := expression.Term{
				Type: expression.Variable,
				Op:   expression.Nop,
				Val:  v}
			if replacement.Contains(toCheck) {
				return false
			}

			expr.Replace(v, &replacement)
		}
	}

	substitution = &sub
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
