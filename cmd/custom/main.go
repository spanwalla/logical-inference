package main

import (
	"fmt"
	"logical-inference/internal/expression"
)

func main() {
	var first = expression.Term{
		Type: expression.Variable,
		Op:   expression.Nop,
		Val:  expression.Value(1),
	}
	var second = expression.Term{
		Type: expression.Variable,
		Op:   expression.Nop,
		Val:  expression.Value(2),
	}
	var third = expression.Term{
		Type: expression.Variable,
		Op:   expression.Nop,
		Val:  expression.Value(3),
	}

	firstExpr := expression.NewExpressionWithTerm(first)
	secondExpr := expression.NewExpressionWithTerm(second)
	thirdExpr := expression.NewExpressionWithTerm(third)

	midExpr := expression.Construct(&firstExpr, expression.Conjunction, &secondExpr)
	expr := expression.Construct(&thirdExpr, expression.Implication, &midExpr)
	fmt.Println(expr.String(), expr)
}
