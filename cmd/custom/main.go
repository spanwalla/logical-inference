package main

import (
	"fmt"
	"logical-inference/internal/expression"
	"logical-inference/internal/logicparser"
	"logical-inference/internal/solver"
)

func main() {
	axioms := []expression.Expression{
		*logicparser.NewExpressionWithString("a>(b>a)"),
		*logicparser.NewExpressionWithString("(a>(b>c))>((a>b)>(a>c))"),
		*logicparser.NewExpressionWithString("(!a>!b)>((!a>b)>a)"),
	}

	target := *logicparser.NewExpressionWithString("a*b>a")
	target.Standardize()
	target.MakeConst()

	slv, err := solver.New(axioms, target, 60000)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer slv.Close()

	err = slv.WriteInitialAxioms()
	if err != nil {
		fmt.Println(err)
		return
	}

	slv.Solve()
	fmt.Println(slv.ThoughtChain())
}
