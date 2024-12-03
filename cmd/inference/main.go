package main

import (
	"fmt"
	"logical-inference/internal/expression"
	"logical-inference/internal/logicparser"
	"logical-inference/internal/solver"
	"time"
)

func main() {
	axioms := []expression.Expression{
		*logicparser.NewExpressionWithString("a>(b>a)"),
		*logicparser.NewExpressionWithString("(a>(b>c))>((a>b)>(a>c))"),
		*logicparser.NewExpressionWithString("(!a>!b)>((!a>b)>a)"),
	}

	var input string
	fmt.Print("Enter expression: ")
	_, err := fmt.Scan(&input)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	target := *logicparser.NewExpressionWithString(input)
	target.Standardize()
	target.MakeConst()

	slv, err := solver.New(axioms, target, 60000)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer slv.Close()

	start := time.Now()
	err = slv.WriteInitialAxioms()
	if err != nil {
		fmt.Println(err)
		return
	}

	slv.Solve()
	duration := time.Since(start)
	fmt.Println(slv.ThoughtChain())
	fmt.Println("Time elapsed:", duration)
}
