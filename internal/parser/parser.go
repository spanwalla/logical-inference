package parser

import (
	"fmt"
	"logical-inference/internal/expression"
	"unicode"
)

type Token int

const (
	Nop Token = iota
	Negation
	Implication
	Disjunction
	Conjunction
	Xor
	Equivalent
	OpenBracket
	CloseBracket
)

var priorities = map[Token]int{
	Nop:          0,
	Negation:     5,
	Implication:  1,
	Disjunction:  3,
	Conjunction:  4,
	Xor:          2,
	Equivalent:   2,
	OpenBracket:  0,
	CloseBracket: 0,
}

var tokenToOperation = map[Token]expression.Operation{
	Nop:         expression.Nop,
	Negation:    expression.Negation,
	Implication: expression.Implication,
	Disjunction: expression.Disjunction,
	Conjunction: expression.Conjunction,
	Xor:         expression.Xor,
	Equivalent:  expression.Equivalent,
}

var opToToken = map[expression.Operation]Token{
	expression.Nop:         Nop,
	expression.Negation:    Negation,
	expression.Implication: Implication,
	expression.Disjunction: Disjunction,
	expression.Conjunction: Conjunction,
	expression.Xor:         Xor,
	expression.Equivalent:  Equivalent,
}

var charToOp = map[rune]expression.Operation{
	'0': expression.Nop,
	'!': expression.Negation,
	'|': expression.Disjunction,
	'*': expression.Conjunction,
	'>': expression.Implication,
	'+': expression.Xor,
	'=': expression.Equivalent,
}

// Parser парсит выражение в список узлов.
type Parser struct {
	brackets   int
	expression string
	operands   []expression.Expression
	operations []Token
}

// NewParser создает новый парсер для логических выражений.
func NewParser(expr string) Parser {
	return Parser{
		brackets:   0,
		expression: expr,
		operands:   []expression.Expression{},
		operations: []Token{},
	}
}

// Parse разбивает выражение на узлы (Nodes).
func (p *Parser) Parse() (expression.Expression, error) {
	lastTokenIsOp := false
	for _, t := range p.expression {
		if unicode.IsSpace(t) {
			continue
		}
		if t == '(' {
			p.operations = append(p.operations, OpenBracket)
			lastTokenIsOp = false
			continue
		}
		if t == ')' {
			if len(p.operations) == 0 {
				return expression.Expression{}, fmt.Errorf("Неправильные скобки")
			}
			for len(p.operations) != 0 && p.operations[len(p.operations)-1] != OpenBracket {
				_ = p.constructNode()
			}
			p.operations = p.operations[:len(p.operations)-1]
			lastTokenIsOp = false
			continue
		}
		if isOperation(t) {
			op := determineOperation(t)
			if op == expression.Negation {
				p.operations = append(p.operations, opToToken[op])
				continue
			}
			if lastTokenIsOp {
				return expression.Expression{}, fmt.Errorf("incorrect input")
			}
			lastTokenIsOp = true
			for len(p.operations) != 0 && priorities[p.operations[len(p.operations)-1]] > priorities[opToToken[op]] {
				_ = p.constructNode()
			}
			p.operations = append(p.operations, opToToken[op])
		} else {
			lastTokenIsOp = false
			p.operands = append(p.operands, expression.NewExpressionWithTerm(determineOperand(t)))
		}
	}
	for len(p.operations) != 0 {
		_ = p.constructNode()
	}
	return p.operands[len(p.operands)-1], nil
}

// constructNode создает новый узел из текущих операций и операндов.
func (p *Parser) constructNode() error {
	if p.operations[len(p.operations)-1] == Negation {

		operand := p.operands[len(p.operands)-1]
		p.operands = p.operands[:len(p.operands)-1]
		p.operations = p.operations[:len(p.operations)-1]

		operand.Negation(0)
		p.operands = append(p.operands, operand)
		return nil

	}

	if len(p.operands) < 2 || len(p.operations) < 1 {
		if p.operations[len(p.operations)-1] == OpenBracket || p.operations[len(p.operations)-1] == CloseBracket {
			return fmt.Errorf("неправильные скобки")
		}
		return fmt.Errorf("что-то пошло не так при формировании узла")
	}

	// Извлекаем узлы
	rhs := p.operands[len(p.operands)-1]
	p.operands = p.operands[:len(p.operands)-1]
	op := p.operations[len(p.operations)-1]
	p.operations = p.operations[:len(p.operations)-1]
	lhs := p.operands[len(p.operands)-1]
	p.operands = p.operands[:len(p.operands)-1]

	p.operands = append(p.operands, expression.Construct(&lhs, tokenToOperation[op], &rhs))
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

func determineOperand(token rune) expression.Term {
	if !('a' <= token && token <= 'z') {
		panic("неправильное имя переменной")
	}
	return expression.Term{
		Type: expression.Variable,
		Op:   expression.Nop,
		Val:  expression.Value(token - 'a' + 1),
	}
}
