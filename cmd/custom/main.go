package main

import (
	"fmt"
	"logical-inference/internal/expression"
	"logical-inference/internal/parser"
	"logical-inference/internal/solver"
)

func main() {
	term := expression.Term{
		expression.Variable,
		expression.Nop,
		expression.Value(1),
	}

	sterm := expression.Term{
		expression.Variable,
		expression.Nop,
		expression.Value(2),
	}

	e := expression.NewExpressionWithTerm(term)
	es := expression.NewExpressionWithTerm(sterm)

	ef := expression.Construct(&es, expression.Implication, &e)

	eg := expression.Construct(&e, expression.Implication, &ef)
	fmt.Println(eg)

	newParsers := []parser.Parser{
		parser.NewParser("a>(b>a)"),
		parser.NewParser("(a>(b>c))>((a>b)>(a>c))"),
		parser.NewParser("(!a>!b)>((!a>b)>a)"),
	}

	axioms := make([]expression.Expression, 0, len(newParsers))
	targetParser := parser.NewParser("a*b>b")
	target, err := targetParser.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, newParser := range newParsers {
		expr, err := newParser.Parse()
		if err != nil {
			fmt.Println(err)
			return
		}
		axioms = append(axioms, expr)
		fmt.Println(expr)
		fmt.Println(expr.String())
	}

	fmt.Println(target)

	slv, err := solver.New(axioms, target, 60000)
	if err != nil {
		fmt.Println(err)
	}
	// slv.Solve()
	// fmt.Println(slv.ThoughtChain())
	defer func(slv solver.Solver) {
		err := slv.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(slv)
}
