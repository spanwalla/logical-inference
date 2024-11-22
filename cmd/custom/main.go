package main

import (
	"fmt"
	"logical-inference/internal/expression"
	"logical-inference/internal/solver"
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

	axioms := []expression.Expression{
		firstExpr,
		midExpr,
		expr,
	}
	slv, err := solver.New(axioms, secondExpr, 60000)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer func(slv *solver.Solver) {
		err := slv.Close()
		if err != nil {

		}
	}(slv)
}
