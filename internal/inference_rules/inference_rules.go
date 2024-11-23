package inferencerules

import (
	"logical-inference/internal/expression"
	"logical-inference/internal/helper"
)

func ApplyModusPonens(lhs expression.Expression, rhs expression.Expression) expression.Expression {
	if lhs.Empty() || rhs.Empty() {
		return expression.Expression{}
	}
	if rhs.Nodes[0].Term.Op != expression.Implication {
		return expression.Expression{}
	}
	//попытка применить унификацию
	substitution := make(map[expression.Value]expression.Expression)
	if !helper.GetUnification(lhs, rhs.CopySubtree(rhs.Subtree(0).Left()), &substitution) {
		return expression.Expression{}
	}
	result := rhs
	result.ChangeVariables(lhs.MaxValue() + 1)
	vars := result.Variables()

	for var_ := range vars {
		if _, exist := substitution[expression.Value(var_)]; !exist {
			continue
		}
		change := substitution[expression.Value(var_)]
		_, exists := substitution[change.Nodes[0].Term.Val]
		for change.Nodes[0].Term.Type == expression.Variable && exists {
			shouldNegate := change.Nodes[0].Term.Op == expression.Negation
			change = substitution[change.Nodes[0].Term.Val]
			if shouldNegate {
				change.Negation(0)
			}
		}
		result.Replace(expression.Value(var_), &change)
	}
	result = result.CopySubtree(result.Subtree(0).Right())
	result.Normalize()
	return result
}
