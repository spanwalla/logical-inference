package main

import (
	"fmt"
	"logical-inference/internal/expression"
	"logical-inference/internal/logicparser"
	"logical-inference/internal/rules"
	"logical-inference/internal/solver"
)

func main() {
	axioms := []expression.Expression{
		*logicparser.NewExpressionWithString("a>(b>a)"),
		*logicparser.NewExpressionWithString("(a>(b>c))>((a>b)>(a>c))"),
		*logicparser.NewExpressionWithString("(!a>!b)>((!a>b)>a)"),
	}

	target := *logicparser.NewExpressionWithString("a*b>a")

	e := rules.ApplyModusPonens(axioms[0], axioms[0])
	e.Normalize()

	slv, err := solver.New(axioms, target, 60000)
	defer func(slv *solver.Solver) {
		err = slv.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(slv)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = slv.WriteInitialAxioms()
	if err != nil {
		fmt.Println(err)
		return
	}
	slv.Solve()
	fmt.Println(slv.ThoughtChain())
}
