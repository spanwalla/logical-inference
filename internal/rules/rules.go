package rules

import (
	"github.com/tiendc/go-deepcopy"
	"logical-inference/internal/expression"
	"logical-inference/internal/helper"
)

func ApplyModusPonens(lhs, rhs expression.Expression) *expression.Expression {
	if lhs.Empty() || rhs.Empty() {
		return expression.NewExpression()
	}

	if rhs.Nodes[0].Term.Op != expression.Implication {
		return expression.NewExpression()
	}

	// Попытка применить унификацию
	substitution := make(map[expression.Value]expression.Expression)
	if !helper.GetUnification(lhs, *rhs.CopySubtree(rhs.Subtree(0).Left()), &substitution) {
		return expression.NewExpression()
	}

	contains := func(key expression.Value) bool {
		_, ok := substitution[key]
		return ok
	}

	var result expression.Expression
	_ = deepcopy.Copy(&result, &rhs)
	result.ChangeVariables(lhs.MaxValue() + 1)
	vars := result.Variables()

	for _, value := range vars {
		if tmp, exists := substitution[value]; exists {
			var change expression.Expression
			_ = deepcopy.Copy(&change, &tmp)
			for change.Nodes[0].Term.Type == expression.Variable && contains(change.Nodes[0].Term.Val) {
				shouldNegate := change.Nodes[0].Term.Op == expression.Negation
				tmp = substitution[change.Nodes[0].Term.Val]
				_ = deepcopy.Copy(&change, &tmp)
				if shouldNegate {
					change.Negation(0)
				}
			}
			result.Replace(value, change)
		}
	}

	r := result.CopySubtree(result.Subtree(0).Right())
	r.Normalize()
	return r
}
