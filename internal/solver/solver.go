package solver

import (
	"fmt"
	"github.com/scylladb/go-set/strset"
	"logical-inference/internal/expression"
	"logical-inference/internal/helper"
	"logical-inference/internal/logicparser"
	"logical-inference/internal/rules"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type Node struct {
	Expression   string   // Строковое представление выражения
	Rule         string   // Правило вывода
	Dependencies []string // Зависимости
}

func NewNode(expression, rule string) *Node {
	if rule == "" {
		rule = "axiom"
	}
	return &Node{
		Expression:   expression,
		Rule:         rule,
		Dependencies: []string{},
	}
}

func msSinceEpoch() uint64 {
	return uint64(time.Now().UnixNano() / int64(time.Millisecond))
}

type Solver struct {
	knownAxioms strset.Set
	axioms      []expression.Expression
	produced    []expression.Expression
	targets     []expression.Expression

	timeLimit uint64

	builder    strings.Builder
	outputFile *os.File
}

func New(axioms []expression.Expression, target expression.Expression, timeLimit uint64) (*Solver, error) {
	if timeLimit < 1 {
		timeLimit = 60000
	}

	if len(axioms) < 3 {
		return &Solver{}, fmt.Errorf("not enough axioms to solve (3 required)")
	}

	file, err := os.Create("conclusions.txt")
	if err != nil {
		return &Solver{}, fmt.Errorf("failed to create file: %w", err)
	}

	return &Solver{
		knownAxioms: *strset.New(),
		axioms:      axioms,
		produced:    []expression.Expression{},
		targets:     []expression.Expression{target},
		timeLimit:   timeLimit,
		builder:     strings.Builder{},
		outputFile:  file,
	}, nil
}

// Close закрывает поток вывода, необходимо использовать всегда.
func (s *Solver) Close() error {
	if s.outputFile != nil {
		return s.outputFile.Close()
	}
	return nil
}

func (s *Solver) WriteInitialAxioms() error {
	axioms := []expression.Expression{
		*logicparser.NewExpressionWithString("a>(b>a)"),
		*logicparser.NewExpressionWithString("(a>(b>c))>((a>b)>(a>c))"),
		*logicparser.NewExpressionWithString("(!a>!b)>((!a>b)>a)"),
	}

	axioms = append(axioms, *rules.ApplyModusPonens(axioms[0], axioms[0]))
	axioms = append(axioms, *rules.ApplyModusPonens(axioms[1], axioms[0]))
	axioms = append(axioms, *rules.ApplyModusPonens(axioms[3], axioms[1]))
	axioms = append(axioms, *rules.ApplyModusPonens(axioms[4], axioms[1]))
	axioms = append(axioms, *rules.ApplyModusPonens(axioms[2], axioms[5]))
	axioms = append(axioms, *rules.ApplyModusPonens(axioms[6], axioms[6]))
	axioms = append(axioms, *rules.ApplyModusPonens(axioms[7], axioms[8]))
	axioms = append(axioms, *rules.ApplyModusPonens(axioms[3], axioms[9]))

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
		if _, err := fmt.Fprintf(s.outputFile, "%s mp %s %s\n", axioms[t[0]].String(), axioms[t[1]].String(),
			axioms[t[2]].String()); err != nil {
			return err
		}
	}
	return nil
}

// Проверка, доказано ли целевое выражение
func (s *Solver) isTargetProvedBy(expr expression.Expression) bool {
	if expr.Empty() {
		return false
	}

	for _, target := range s.targets {
		if helper.IsEqual(target, expr) {
			return true
		}
	}
	return false
}

// Проверка, является ли выражение хорошим
func (s *Solver) isGoodExpression(expr expression.Expression, maxLen int) bool {
	return !(expr.Size() > maxLen || expr.Empty() ||
		expr.Nodes[0].Term.Op == expression.Conjunction ||
		expr.Operations(expression.Conjunction) > 1)
}

func (s *Solver) deductionTheoremDecomposition(expr expression.Expression) bool {
	if expr.Empty() {
		return false
	}

	if expr.Nodes[0].Term.Op != expression.Implication {
		return false
	}

	// Γ ⊢ A → B <=> Γ U {A} ⊢ B
	s.axioms = append(s.axioms, *expr.CopySubtree(expr.Subtree(0).Left()))
	s.targets = append(s.targets, *expr.CopySubtree(expr.Subtree(0).Right()))
	return true
}

func (s *Solver) produce(maxLen int) {
	if len(s.produced) == 0 {
		return
	}

	newlyProduced := make([]expression.Expression, 0, len(s.produced)*2)

	for _, expr := range s.produced {
		// Проверка времени
		if msSinceEpoch() > s.timeLimit {
			break
		}

		// Пропустить слишком длинные выражения
		if expr.Size() > maxLen {
			continue
		}

		// Нормализовать и добавить выражение к аксиомам
		expr.Normalize()
		s.axioms = append(s.axioms, expr)

		// Проверить, доказано ли целевое выражение
		if s.isTargetProvedBy(s.axioms[len(s.axioms)-1]) {
			return
		}

		// Создать новые выражения через modus-ponens
		for j := range s.axioms {
			newExpr := *rules.ApplyModusPonens(s.axioms[j], s.axioms[len(s.axioms)-1])

			if !s.isGoodExpression(newExpr, maxLen) || s.knownAxioms.Has(newExpr.String()) {
				continue
			}

			newlyProduced = append(newlyProduced, newExpr)
			s.knownAxioms.Add(newExpr.String())

			_, err := fmt.Fprintf(s.outputFile, "%s mp %s %s\n", newExpr.String(), s.axioms[j].String(), s.axioms[len(s.axioms)-1].String())
			if err != nil {
				fmt.Println(err)
				return
			}

			if s.isTargetProvedBy(newExpr) {
				s.axioms = append(s.axioms, newlyProduced[len(newlyProduced)-1])
				return
			}

			if len(s.axioms) == j+1 {
				break
			}

			// Инверсный порядок modus ponens
			newExpr = *rules.ApplyModusPonens(s.axioms[len(s.axioms)-1], s.axioms[j])

			if !s.isGoodExpression(newExpr, maxLen) || s.knownAxioms.Has(newExpr.String()) {
				continue
			}

			newlyProduced = append(newlyProduced, newExpr)
			s.knownAxioms.Add(newExpr.String())

			_, err = fmt.Fprintf(s.outputFile, "%s mp %s %s\n", newExpr.String(), s.axioms[len(s.axioms)-1].String(), s.axioms[j].String())
			if err != nil {
				fmt.Println(err)
				return
			}

			if s.isTargetProvedBy(newlyProduced[len(newlyProduced)-1]) {
				s.axioms = append(s.axioms, newlyProduced[len(newlyProduced)-1])
				return
			}
		}
	}

	// Проверка времени
	if msSinceEpoch() > s.timeLimit {
		return
	}

	// Сортируем новые выражения по длине
	sort.Slice(newlyProduced, func(i, j int) bool {
		return newlyProduced[i].Size() < newlyProduced[j].Size()
	})

	s.produced = newlyProduced
}

func (s *Solver) Solve() {
	s.builder.Reset()
	limit := 20

	for s.deductionTheoremDecomposition(s.targets[len(s.targets)-1]) {
		prev := s.targets[len(s.targets)-2]
		curr := s.targets[len(s.targets)-1]
		axiom := s.axioms[len(s.axioms)-1]

		s.builder.WriteString(fmt.Sprintf("deduction theorem: Γ ⊢ %s <=> Γ U {%s} ⊢ %s\n", prev.String(), axiom.String(), curr.String()))
	}

	for i := 0; i < len(s.axioms); i++ {
		s.axioms[i].Normalize()
		s.produced = append(s.produced, s.axioms[i])

		_, err := fmt.Fprintf(s.outputFile, "%s axiom\n", s.axioms[i].String())
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// isr rule
	s.produced = append(s.produced, *logicparser.NewExpressionWithString("(!a>!b)>(b>a)"))
	s.axioms = make([]expression.Expression, 0)
	s.knownAxioms = *strset.New()

	// calculate the stopping criterion
	now := msSinceEpoch()
	s.timeLimit = now + s.timeLimit
	if now > math.MaxUint64-s.timeLimit {
		s.timeLimit = math.MaxUint64
	}

	for msSinceEpoch() < s.timeLimit {
		s.produce(limit)
		if s.isTargetProvedBy(s.axioms[len(s.axioms)-1]) {
			break
		}
	}

	found := false
	for _, expr := range s.axioms {
		if s.isTargetProvedBy(expr) {
			found = true
			break
		}
	}

	if !found {
		s.builder.WriteString("No proof was found in the time allotted\n")
		return
	}

	proof := *expression.NewExpression()
	targetProved := *expression.NewExpression()

	for _, axiom := range s.axioms {
		if !proof.Empty() {
			break
		}

		for _, target := range s.targets {
			if helper.IsEqual(target, axiom) {
				proof = axiom
				targetProved = target
				break
			}
		}
	}

	s.buildThoughtChain(proof, targetProved)
}

func (s *Solver) buildThoughtChain(proof expression.Expression, provedTarget expression.Expression) {
	/* file, err := os.Open("conclusions.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	conclusions_ := make(map[string]Node)
	indices := make(map[string]int)
	processedProofs := make(map[string]bool)
	chain := make(map[int]Node)
	nextIndex := 1

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text() // Read the current line
		// Split the line into parts
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue // Skip if the line doesn't contain at least an expression and a rule
		}

		expression_ := parts[0]
		rule := parts[1]

		// Skip if the expression already exists in the map
		if _, exists := conclusions_[expression_]; exists {
			continue
		}

		conclusions_[expression_] = Node{
			Expression_: expression_,
			Rule_:       rule,
		}

		// We extract the Node from the map, update the dependencies and save them back
		node := conclusions_[expression_]
		for _, dependency := range parts[2:] {
			node.Dependencies_ = append(node.Dependencies_, dependency)
		}
		// Saving the updated Node back to the map
		conclusions_[expression_] = node
	}

	var treeLevels [][]string
	treeLevels = append(treeLevels, []string{proof.String()})

	for len(treeLevels) > 0 && len(treeLevels[len(treeLevels)-1]) > 0 {
		var level []string

		for _, expression_ := range treeLevels[len(treeLevels)-1] {
			node := conclusions_[expression_]

			if _, exists := processedProofs[node.Expression_]; exists {
				continue
			}

			if node.Rule_ == "axiom" {
				if _, exists := indices[node.Expression_]; exists {
					continue
				}

				chain[nextIndex] = node
				indices[node.Expression_] = nextIndex
				nextIndex++
			}
			// Add dependencies
			for _, dependency := range node.Dependencies_ {
				level = append(level, dependency)
			}
			// Add express in processedProofs
			processedProofs[node.Expression_] = false
		}
		treeLevels = append(treeLevels, level)
	}
	// Reverse the treeLevels slice
	for i, j := 0, len(treeLevels)-1; i < j; i, j = i+1, j-1 {
		treeLevels[i], treeLevels[j] = treeLevels[j], treeLevels[i]
	}
	// Iterate over the reversed treeLevels
	for _, level := range treeLevels {
		for _, expression_ := range level {
			// If the expression is already indexed, continue
			if _, exists := indices[expression_]; exists {
				continue
			}
			// Add the node to the chain
			chain[nextIndex] = conclusions_[expression_]
			indices[expression_] = nextIndex
			nextIndex++
		}
	}

	//var sb strings.Builder

	for i := 1; i < nextIndex; i++ {
		node := chain[i]
		// Add index and point
		s.builder.WriteString(fmt.Sprintf("%d. ", i))
		// fmt.Fprintf(&s.ss, "%d. ", i)

		if node.Rule_ == "axiom" {
			s.builder.WriteString("axiom")
		} else {
			s.builder.WriteString(node.Rule_ + "(")

			// Add dependencies
			for k, dependency := range node.Dependencies_ {
				s.builder.WriteString(fmt.Sprintf("%d", indices[dependency]))

				// Add a comma if this is not the last dependency
				if k+1 != len(node.Dependencies_) {
					s.builder.WriteString(",")
				}
			}
			s.builder.WriteString(")")
		}

		s.builder.WriteString(fmt.Sprintf(": %s\n", node.Expression_))
	}

	substitution := make(map[expression.Value]expression.Expression)

	helper.GetUnification(provedTarget, proof, &substitution)

	if len(substitution) == 0 {
		return
	}

	s.builder.WriteString(fmt.Sprintf("change variables: %s\n", proof.String()))
	for v, st := range substitution {
		s.builder.WriteString(fmt.Sprintf("%c -> %s\n", rune(v+'A'-1), st.String())) // Преобразуем значение в символ
	}

	s.builder.WriteString(fmt.Sprintf("proved: %s\n", provedTarget.String())) */
}

func (s *Solver) ThoughtChain() string {
	return s.builder.String()
}
