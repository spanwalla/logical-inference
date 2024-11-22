package solver

import (
	"fmt"
	"github.com/scylladb/go-set/strset"
	"logical-inference/internal/expression"
	"logical-inference/internal/rules"
	"os"
	"strings"
)

type Solver struct {
	knownAxioms strset.Set
	axioms      []expression.Expression
	produced    []expression.Expression
	targets     []expression.Expression
	// timeLimit
	builder    strings.Builder
	outputFile *os.File
}

func New(axioms []expression.Expression, target expression.Expression, timeLimit uint64) (*Solver, error) {
	// TODO: timeLimit доделать с использованием встроенных штук
	if len(axioms) < 3 {
		return nil, fmt.Errorf("not enough axioms to solve (3 required)")
	}

	file, err := os.Create("conclusions.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return &Solver{
		axioms:     axioms,
		produced:   []expression.Expression{},
		targets:    []expression.Expression{target},
		outputFile: file,
	}, nil
}

func (s *Solver) Close() error {
	// Функция закрывает поток вывода, необходимо использовать всегда.
	if s.outputFile != nil {
		return s.outputFile.Close()
	}
	return nil
}

func (s *Solver) WriteInitialAxioms() error {
	axioms := []expression.Expression{
		expression.NewExpressionWithStringExpression("a>(b>a)"),
		expression.NewExpressionWithStringExpression("(a>(b>c))>((a>b)>(a>c))"),
		expression.NewExpressionWithStringExpression("(!a>!b)>((!a>b)>a)"),
	}

	axioms = append(axioms, rules.ApplyModusPonens(axioms[0], axioms[0]))
	axioms = append(axioms, rules.ApplyModusPonens(axioms[1], axioms[0]))
	axioms = append(axioms, rules.ApplyModusPonens(axioms[3], axioms[1]))
	axioms = append(axioms, rules.ApplyModusPonens(axioms[4], axioms[1]))
	axioms = append(axioms, rules.ApplyModusPonens(axioms[2], axioms[5]))
	axioms = append(axioms, rules.ApplyModusPonens(axioms[6], axioms[6]))
	axioms = append(axioms, rules.ApplyModusPonens(axioms[7], axioms[8]))
	axioms = append(axioms, rules.ApplyModusPonens(axioms[3], axioms[9]))

	trios := [][3]int{
		{3, 0, 0},
		{4, 1, 0},
		{5, 3, 1},
		{6, 4, 1},
		{7, 2, 5},
		{8, 6, 5},
		{9, 7, 8},
		{10, 3, 9},
	}

	for _, t := range trios {
		if _, err := fmt.Fprintln(s.outputFile, axioms[t[0]], " mp ", axioms[t[1]], " ", axioms[t[2]]); err != nil {
			return err
		}
	}
	return nil
}
