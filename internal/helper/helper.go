package helper

import (
	"fmt"
	"github.com/spanwalla/logical-inference/internal/expression"
	"github.com/spanwalla/logical-inference/internal/pkg/queue"
	"github.com/spanwalla/logical-inference/internal/pkg/stack"
	"github.com/tiendc/go-deepcopy"
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

	// Проходим по всем узлам графа
	for i := expression.Value(0); i < size; i++ {
		if !visited[i] {
			topologicalSortUtil(i, adj, visited, s)
		}
	}

	order := make([]expression.Value, 0, size)
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

	var substCopy expression.Expression
	if err := deepcopy.Copy(&substCopy, &substitution); err != nil {
		fmt.Println("Error deepcopying substitution:", err)
	}
	sub[term.Val] = substCopy
	return true
}

func IsEqual(left, right expression.Expression) bool {
	if left.Size() != right.Size() {
		return false
	}

	if left.Nodes[0].Term.Op != right.Nodes[0].Term.Op {
		return false
	}

	var leftCopy, rightCopy expression.Expression
	_ = deepcopy.Copy(&leftCopy, &left)
	_ = deepcopy.Copy(&rightCopy, &right)
	leftCopy.Normalize()
	rightCopy.Normalize()

	return leftCopy.Equals(rightCopy, true)
}

func GetUnification(left, right expression.Expression, substitution *map[expression.Value]expression.Expression) bool {
	sub := make(map[expression.Value]expression.Expression)
	subContains := func(key expression.Value) bool {
		_, ok := sub[key]
		return ok
	}

	var leftCopy, rightCopy expression.Expression
	_ = deepcopy.Copy(&leftCopy, &left)
	_ = deepcopy.Copy(&rightCopy, &right)
	rightCopy.ChangeVariables(leftCopy.MaxValue() + 1)
	v := rightCopy.MaxValue() + 1

	q := queue.New[[2]uint]()
	q.Push([2]uint{leftCopy.Subtree(0).Self(), rightCopy.Subtree(0).Self()})

	var lhs, rhs expression.Expression

	for !q.Empty() {
		el := q.Pop()
		leftIdx, rightIdx := el[0], el[1]
		leftTerm, rightTerm := leftCopy.Nodes[leftIdx].Term, rightCopy.Nodes[rightIdx].Term

		// case 0: both terms are functions
		if leftTerm.Type == expression.Function && rightTerm.Type == expression.Function {
			if leftTerm.Op != rightTerm.Op {
				return false
			}

			q.Push([2]uint{leftCopy.Subtree(leftIdx).Left(), rightCopy.Subtree(rightIdx).Left()})
			q.Push([2]uint{leftCopy.Subtree(leftIdx).Right(), rightCopy.Subtree(rightIdx).Right()})
			continue
		}

		lhs = *leftCopy.CopySubtree(leftIdx)
		rhs = *rightCopy.CopySubtree(rightIdx)

		for lhs.Nodes[0].Term.Type == expression.Variable && subContains(lhs.Nodes[0].Term.Val) {
			shouldNegate := lhs.Nodes[0].Term.Op == expression.Negation
			tmp := sub[lhs.Nodes[0].Term.Val]
			_ = deepcopy.Copy(&lhs, &tmp)
			if shouldNegate {
				lhs.Negation(0)
			}
		}
		for rhs.Nodes[0].Term.Type == expression.Variable && subContains(rhs.Nodes[0].Term.Val) {
			shouldNegate := rhs.Nodes[0].Term.Op == expression.Negation
			tmp := sub[rhs.Nodes[0].Term.Val]
			_ = deepcopy.Copy(&rhs, &tmp)
			if shouldNegate {
				rhs.Negation(0)
			}
		}

		// case 1: both terms are constants
		if lhs.Nodes[0].Term.Type == expression.Constant && rhs.Nodes[0].Term.Type == expression.Constant {
			if lhs.Nodes[0].Term != rhs.Nodes[0].Term {
				return false
			}
			continue
		}

		// case 2: left term is constant and right is variable
		if lhs.Nodes[0].Term.Type == expression.Constant && rhs.Nodes[0].Term.Type == expression.Variable {
			if rhs.Nodes[0].Term.Op == expression.Negation {
				if lhs.Nodes[0].Term.Op != expression.Negation {
					lhs.Nodes[0].Term.Op = expression.Negation
				} else {
					lhs.Nodes[0].Term.Op = expression.Nop
				}
			}

			if !AddConstraint(rhs.Nodes[0].Term, lhs, sub) {
				return false
			}
			continue
		}

		// case 3: left term is variable and right is constant
		if lhs.Nodes[0].Term.Type == expression.Variable && rhs.Nodes[0].Term.Type == expression.Constant {
			if lhs.Nodes[0].Term.Op == expression.Negation {
				if rhs.Nodes[0].Term.Op != expression.Negation {
					rhs.Nodes[0].Term.Op = expression.Negation
				} else {
					rhs.Nodes[0].Term.Op = expression.Nop
				}
			}

			if !AddConstraint(lhs.Nodes[0].Term, rhs, sub) {
				return false
			}
			continue
		}

		// case 4: both terms are variables
		if lhs.Nodes[0].Term.Type == expression.Variable && rhs.Nodes[0].Term.Type == expression.Variable {
			if lhs.Nodes[0].Term.Val == rhs.Nodes[0].Term.Val {
				if lhs.Nodes[0].Term.Op != rhs.Nodes[0].Term.Op {
					return false
				}
				continue
			}

			op := expression.Nop
			if lhs.Nodes[0].Term.Op == expression.Negation || rhs.Nodes[0].Term.Op == expression.Negation {
				op = expression.Negation
			}
			expr := *expression.NewExpressionWithTerm(expression.Term{
				Type: expression.Variable,
				Op:   op,
				Val:  v,
			})
			v++
			var negExpr expression.Expression
			_ = deepcopy.Copy(&negExpr, &expr)
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

		// case 5: left term is a function
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

		// case 6: right term is a function
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
		order[i] += 1

		if !subContains(order[i]) {
			continue
		}

		var expr expression.Expression
		tmp := sub[order[i]]
		_ = deepcopy.Copy(&expr, &tmp)
		if expr.Nodes[0].Term.Type != expression.Function {
			continue
		}

		for _, w := range expr.Variables() {
			if !subContains(w) {
				continue
			}

			var replacement expression.Expression
			tmp = sub[w]
			_ = deepcopy.Copy(&replacement, &tmp)
			for replacement.Nodes[0].Term.Type == expression.Variable && subContains(replacement.Nodes[0].Term.Val) {
				shouldNegate := replacement.Nodes[0].Term.Op == expression.Negation
				tmp = sub[replacement.Nodes[0].Term.Val]
				_ = deepcopy.Copy(&replacement, &tmp)
				if shouldNegate {
					replacement.Negation(0)
				}
			}

			toCheck := expression.Term{
				Type: expression.Variable,
				Op:   expression.Nop,
				Val:  w,
			}
			if replacement.Contains(toCheck) {
				return false
			}

			expr.Replace(w, replacement)
			sub[order[i]] = expr
		}
	}

	*substitution = sub
	return true
}
