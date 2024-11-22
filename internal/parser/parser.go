package parser

import (
	"fmt"
	"strings"
	"unicode"

	"logical-inference/internal/expression"
)

// Node представляет узел: Term и его связи.
type Node expression.Node

// Parser парсит выражение в список узлов.
type Parser struct {
	expression string
	nodes      []Node
	operations []expression.Operation
}

// NewParser создает новый парсер для логических выражений.
func NewParser(expr string) *Parser {
	return &Parser{
		expression: strings.TrimSpace(expr),
		nodes:      []Node{},
		operations: []expression.Operation{},
	}
}

// Parse разбивает выражение на узлы (Nodes).
func (p *Parser) Parse() ([]Node, error) {
	//fmt.Println("enter parse")
	var currentIndex int
	fmt.Println(currentIndex)
	lastTokenIsOp := false

	for _, token := range p.expression {
		if unicode.IsSpace(token) {
			continue
		}

		switch {
		case token == '(':
			p.operations = append(p.operations, expression.Nop) // Открывающая скобка как заглушка
			lastTokenIsOp = false
		case token == ')':
			for len(p.operations) > 0 && p.operations[len(p.operations)-1] != expression.Nop {
				if err := p.constructNode(&currentIndex); err != nil {
					return nil, err
				}
			}
			if len(p.operations) > 0 {
				p.operations = p.operations[:len(p.operations)-1] // Убираем скобку
			}
			lastTokenIsOp = false
		case isOperation(token):
			op := determineOperation(token)
			if lastTokenIsOp && op != expression.Negation {
				return nil, fmt.Errorf("ошибка: два оператора подряд")
			}
			lastTokenIsOp = true
			for len(p.operations) > 0 && priority(p.operations[len(p.operations)-1]) >= priority(op) {
				if err := p.constructNode(&currentIndex); err != nil {
					return nil, err
				}
			}
			p.operations = append(p.operations, op)
		default:
			//fmt.Println("Space found2")
			if !unicode.IsLetter(token) {
				return nil, fmt.Errorf("неверный символ: %c", token)
			}
			lastTokenIsOp = false
			term := expression.Term{
				Type: expression.Variable,
				Op:   expression.Nop,
				Val:  expression.Value(token - 'a' + 1),
			}
			p.nodes = append(p.nodes, Node{
				Term: term,
				Rel:  expression.NewSelfRelation(currentIndex),
			})
			currentIndex++
		}
	}

	for len(p.operations) > 0 {
		if err := p.constructNode(&currentIndex); err != nil {
			return nil, err
		}
	}

	return p.nodes, nil
}

// constructNode создает новый узел из текущих операций и операндов.
func (p *Parser) constructNode(currentIndex *int) error {
	if len(p.operations) == 0 || len(p.nodes) < 1 {
		return fmt.Errorf("недостаточно данных для создания узла")
	}

	op := p.operations[len(p.operations)-1]
	p.operations = p.operations[:len(p.operations)-1] // Убираем операцию

	if op == expression.Negation {
		// Унарная операция
		node := &p.nodes[len(p.nodes)-1]
		node.Term.Op = expression.Negation
		return nil
	}

	if len(p.nodes) < 2 {
		return fmt.Errorf("недостаточно операндов для бинарной операции")
	}

	// Бинарная операция
	right := p.nodes[len(p.nodes)-1]
	p.nodes = p.nodes[:len(p.nodes)-1] // Убираем правый операнд
	left := p.nodes[len(p.nodes)-1]
	p.nodes = p.nodes[:len(p.nodes)-1] // Убираем левый операнд

	newNode := Node{
		Term: expression.Term{
			Type: expression.Function,
			Op:   op,
			Val:  0,
		},
		Rel: expression.NewRelationWithIndices(*currentIndex, left.Rel.Self(), right.Rel.Self(), -1),
	}

	p.nodes = append(p.nodes, newNode)
	*currentIndex++
	return nil
}

// isOperation проверяет, является ли символ оператором.
func isOperation(token rune) bool {
	_, ok := charToOp[token]
	return ok
}

// determineOperation определяет операцию из символа.
func determineOperation(token rune) expression.Operation {
	if op, ok := charToOp[token]; ok {

		return op
	}
	return expression.Nop
}

// priority возвращает приоритет операции.
func priority(op expression.Operation) int {
	if pr, ok := priorities[op]; ok {
		fmt.Print(op, pr)
		return pr
	}
	return 0
}

var charToOp = map[rune]expression.Operation{
	'!': expression.Negation,
	'|': expression.Disjunction,
	'*': expression.Conjunction,
	'>': expression.Implication,
	'+': expression.Xor,
	'=': expression.Equivalent,
}

var priorities = map[expression.Operation]int{
	expression.Nop:         0,
	expression.Negation:    5,
	expression.Conjunction: 3,
	expression.Disjunction: 1,
	expression.Xor:         2,
	expression.Implication: 4,
	expression.Equivalent:  2,
}
