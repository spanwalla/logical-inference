package main

import (
	"fmt"
	"logical-inference/internal/expression"
	"logical-inference/internal/parser"
	"logical-inference/internal/solver"
)

func main() {
	axiomParsers := []parser.Parser{
		parser.NewParser("a>(b>a)"),
		parser.NewParser("(a>(b>c))>((a>b)>(a>c))"),
		parser.NewParser("(!a>!b)>((!a>b)>a)"),
	}

	axioms := make([]expression.Expression, 0, len(axiomParsers))
	targetParser := parser.NewParser("a*b>a")
	target, err := targetParser.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, axiomParser := range axiomParsers {
		expr, err := axiomParser.Parse()
		if err != nil {
			fmt.Println(err)
			return
		}
		axioms = append(axioms, expr)
	}

	slv, err := solver.New(axioms, target, 60000)
	if err != nil {
		fmt.Println(err)
	}
	slv.Solve()
	fmt.Println(slv.ThoughtChain())
	defer func(slv solver.Solver) {
		err := slv.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(slv)

	fmt.Println(target.Variables())
	target.MakeConst()
	fmt.Println(target.Operations(expression.Conjunction))
}
