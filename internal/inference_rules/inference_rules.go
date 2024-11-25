package inferencerules

import (
	"logical-inference/internal/expression"
	"logical-inference/internal/helper"
)

func ApplyModusPonens(lhs expression.Expression, rhs expression.Expression) expression.Expression {
	if lhs.Empty() || rhs.Empty() {
		return expression.NewExpression()
	}
	if rhs.Nodes[0].Term.Op != expression.Implication {
		return expression.NewExpression()
	}
	// Попытка применить унификацию
	substitution := make(map[expression.Value]expression.Expression)
	if !helper.GetUnification(lhs, rhs.CopySubtree(rhs.Subtree(0).Left()), &substitution) {
		return expression.NewExpression()
	}
	result := rhs
	result.ChangeVariables(lhs.MaxValue() + 1)
	vars := result.Variables()

	for _, var_ := range vars {
		if _, exist := substitution[var_]; !exist {
			continue
		}
		change := substitution[var_]
		for change.Nodes[0].Term.Type == expression.Variable {
			if _, exists := substitution[change.Nodes[0].Term.Val]; !exists {
				break
			}

			shouldNegate := change.Nodes[0].Term.Op == expression.Negation
			change = substitution[change.Nodes[0].Term.Val]
			if shouldNegate {
				change.Negation(0)
			}
		}
		result.Replace(var_, &change)
	}

	result = result.CopySubtree(result.Subtree(0).Right())
	result.Normalize()
	return result
}
