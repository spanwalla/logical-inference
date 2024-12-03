package solver

import (
	"bufio"
	"fmt"
	"github.com/scylladb/go-set/strset"
	"github.com/spanwalla/logical-inference/internal/expression"
	"github.com/spanwalla/logical-inference/internal/helper"
	"github.com/spanwalla/logical-inference/internal/logicparser"
	"github.com/spanwalla/logical-inference/internal/pkg/alphabet"
	"github.com/spanwalla/logical-inference/internal/rules"
	"github.com/tiendc/go-deepcopy"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Node struct {
	Expression   string   // Строковое представление выражения
	Rule         string   // Правило вывода
	Dependencies []string // Зависимости
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
	fileWriter *bufio.Writer
}

func New(axioms []expression.Expression, target expression.Expression, timeLimit uint64) (*Solver, error) {
	if timeLimit < 1 {
		timeLimit = 60000
	}

	if len(axioms) < 3 {
		return nil, fmt.Errorf("not enough axioms to solve (3 required)")
	}

	file, err := os.Create("conclusions.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	var targetCopy expression.Expression
	_ = deepcopy.Copy(&targetCopy, &target)

	return &Solver{
		knownAxioms: *strset.New(),
		axioms:      axioms,
		produced:    []expression.Expression{},
		targets:     []expression.Expression{targetCopy},
		timeLimit:   timeLimit,
		builder:     strings.Builder{},
		outputFile:  file,
		fileWriter:  bufio.NewWriter(file),
	}, nil
}

// Close закрывает поток вывода, необходимо использовать всегда.
func (s *Solver) Close() {
	if s.outputFile != nil {
		err := s.outputFile.Close()
		if err != nil {
			fmt.Println("failed to close output file:", err)
		}
	}
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
		if _, err := fmt.Fprintf(s.fileWriter, "%s mp %s %s\n", axioms[t[0]].String(), axioms[t[1]].String(),
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
	var expr expression.Expression

	for i := range s.produced {
		if msSinceEpoch() > s.timeLimit {
			break
		}

		if s.produced[i].Size() > maxLen {
			continue
		}

		s.produced[i].Normalize()
		tmp := s.produced[i]
		var copiedTmp expression.Expression
		_ = deepcopy.Copy(&copiedTmp, &tmp)
		s.axioms = append(s.axioms, copiedTmp)

		if s.isTargetProvedBy(copiedTmp) {
			return
		}

		for j := 0; j < len(s.axioms); j++ {
			expr = *rules.ApplyModusPonens(s.axioms[j], s.axioms[len(s.axioms)-1])

			if !s.isGoodExpression(expr, maxLen) || s.knownAxioms.Has(expr.String()) {
				continue
			}

			_ = deepcopy.Copy(&tmp, &expr)
			newlyProduced = append(newlyProduced, tmp)
			s.knownAxioms.Add(tmp.String())

			_, err := fmt.Fprintf(s.fileWriter, "%s mp %s %s\n", tmp.String(), s.axioms[j].String(), s.axioms[len(s.axioms)-1].String())
			if err != nil {
				fmt.Println(err)
				return
			}

			if s.isTargetProvedBy(tmp) {
				var axiom expression.Expression
				_ = deepcopy.Copy(&axiom, &tmp)
				s.axioms = append(s.axioms, axiom)
				return
			}

			if len(s.axioms) == j+1 {
				break
			}

			// Обратный порядок
			expr = *rules.ApplyModusPonens(s.axioms[len(s.axioms)-1], s.axioms[j])

			if !s.isGoodExpression(expr, maxLen) || s.knownAxioms.Has(expr.String()) {
				continue
			}

			_ = deepcopy.Copy(&tmp, &expr)
			newlyProduced = append(newlyProduced, tmp)
			s.knownAxioms.Add(tmp.String())

			_, err = fmt.Fprintf(s.fileWriter, "%s mp %s %s\n", tmp.String(), s.axioms[len(s.axioms)-1].String(), s.axioms[j].String())
			if err != nil {
				fmt.Println(err)
				return
			}

			if s.isTargetProvedBy(tmp) {
				var axiom expression.Expression
				_ = deepcopy.Copy(&axiom, &tmp)
				s.axioms = append(s.axioms, axiom)
				return
			}
		}
	}

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

	for i := range s.axioms {
		s.axioms[i].Normalize()

		tmp := s.axioms[i]
		var copiedTmp expression.Expression
		_ = deepcopy.Copy(&copiedTmp, &tmp)
		s.produced = append(s.produced, copiedTmp)

		_, err := fmt.Fprintf(s.fileWriter, "%s axiom\n", s.axioms[i].String())
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
	if now > math.MaxUint64-s.timeLimit {
		s.timeLimit = math.MaxUint64
	} else {
		s.timeLimit = now + s.timeLimit
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
				_ = deepcopy.Copy(&proof, &axiom)
				_ = deepcopy.Copy(&targetProved, &target)
				break
			}
		}
	}

	if err := s.fileWriter.Flush(); err != nil {
		fmt.Println("Error flushing writer:", err)
		return
	}
	s.buildThoughtChain(proof, targetProved)
}

func (s *Solver) buildThoughtChain(proof expression.Expression, provedTarget expression.Expression) {
	if _, err := s.outputFile.Seek(0, 0); err != nil {
		fmt.Println("Error seeking file:", err)
		return
	}

	conclusions := make(map[string]Node)
	indices := make(map[string]int)
	chain := make(map[int]Node)
	processedProofs := *strset.New()
	nextIndex := 1

	scanner := bufio.NewScanner(s.outputFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue // Skip if the line doesn't contain at least an expression and a rule
		}

		expr := parts[0]
		rule := parts[1]

		// Skip if the expression already exists in the map
		if _, exists := conclusions[expr]; exists {
			continue
		}

		conclusions[expr] = Node{
			Expression:   expr,
			Rule:         rule,
			Dependencies: parts[2:],
		}
	}

	var treeLevels [][]string
	treeLevels = append(treeLevels, []string{proof.String()})

	for len(treeLevels) > 0 && len(treeLevels[len(treeLevels)-1]) > 0 {
		level := make([]string, 0)
		for _, expr := range treeLevels[len(treeLevels)-1] {
			node := conclusions[expr]
			if processedProofs.Has(node.Expression) {
				continue
			}

			if node.Rule == "axiom" {
				if _, exists := indices[node.Expression]; exists {
					continue
				}

				chain[nextIndex] = node
				indices[node.Expression] = nextIndex
				nextIndex++
			}

			for _, dep := range node.Dependencies {
				level = append(level, dep)
			}

			processedProofs.Add(node.Expression)
		}
		treeLevels = append(treeLevels, level)
	}

	// Reverse the treeLevels slice
	for i, j := 0, len(treeLevels)-1; i < j; i, j = i+1, j-1 {
		treeLevels[i], treeLevels[j] = treeLevels[j], treeLevels[i]
	}

	for _, level := range treeLevels {
		for _, expr := range level {
			if _, exists := indices[expr]; exists {
				continue
			}

			chain[nextIndex] = conclusions[expr]
			indices[expr] = nextIndex
			nextIndex++
		}
	}

	for i := 1; i < nextIndex; i++ {
		node := chain[i]
		s.builder.WriteString(fmt.Sprintf("%d. ", i))

		if node.Rule == "axiom" {
			s.builder.WriteString("axiom")
		} else {
			s.builder.WriteString(fmt.Sprintf("%s(", node.Rule))
			for k := 0; k < len(node.Dependencies); k++ {
				s.builder.WriteString(strconv.Itoa(indices[node.Dependencies[k]]))

				if len(node.Dependencies) != k+1 {
					s.builder.WriteString(",")
				}
			}
			s.builder.WriteString(")")
		}
		s.builder.WriteString(fmt.Sprintf(": %s\n", node.Expression))
	}

	// Change variables if required
	substitution := make(map[expression.Value]expression.Expression)
	helper.GetUnification(provedTarget, proof, &substitution)
	if len(substitution) == 0 {
		return
	}

	s.builder.WriteString(fmt.Sprintf("change variables: %s\n", proof.String()))
	for key, value := range substitution {
		letter, err := alphabet.GetLetter(int(key), true)
		if err != nil {
			fmt.Println("Error in substitution, letter set to 'X':", err)
			letter = 'X'
		}
		s.builder.WriteString(fmt.Sprintf("%c \u2192 %s\n", letter, value.String()))
	}
	s.builder.WriteString(fmt.Sprintf("proved: %s\n", provedTarget.String()))
}

func (s *Solver) ThoughtChain() string {
	return s.builder.String()
}
