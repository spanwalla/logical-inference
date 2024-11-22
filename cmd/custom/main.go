package main

import (
	"fmt"
	"logical-inference/internal/logic"
)

func main() {
	var first = logic.Term{
		Type: logic.Variable,
		Op:   logic.Nop,
		Val:  logic.Value(1),
	}
	var second = logic.Term{
		Type: logic.Variable,
		Op:   logic.Nop,
		Val:  logic.Value(2),
	}
	var third = logic.Term{
		Type: logic.Variable,
		Op:   logic.Nop,
		Val:  logic.Value(3),
	}

	firstExpr := logic.NewExpressionWithTerm(first)
	secondExpr := logic.NewExpressionWithTerm(second)
	thirdExpr := logic.NewExpressionWithTerm(third)

	midExpr := logic.Construct(&firstExpr, logic.Conjunction, &secondExpr)
	expr := logic.Construct(&thirdExpr, logic.Implication, &midExpr)
	fmt.Println(expr.String(), expr)
}
