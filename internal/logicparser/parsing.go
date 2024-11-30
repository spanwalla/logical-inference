package logicparser

import (
	"fmt"
	"logical-inference/internal/expression"
	"logical-inference/internal/pkg/stack"
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

var priority = map[Token]int{
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
	'\x00': expression.Nop,
	'!':    expression.Negation,
	'|':    expression.Disjunction,
	'*':    expression.Conjunction,
	'>':    expression.Implication,
	'+':    expression.Xor,
	'=':    expression.Equivalent,
}

// LogicParser парсит выражение в список узлов.
type LogicParser struct {
	brackets   int
	expression string
	operands   *stack.Stack[expression.Expression]
	operations *stack.Stack[Token]
}

// NewLogicParser создает новый анализатор для логических выражений.
func NewLogicParser(expr string) LogicParser {
	return LogicParser{
		brackets:   0,
		expression: expr,
		operands:   stack.New[expression.Expression](),
		operations: stack.New[Token](),
	}
}

// Parse разбивает выражение на узлы (Nodes).
func (p *LogicParser) Parse() (*expression.Expression, error) {
	lastTokenIsOp := false
	for _, t := range p.expression {
		if unicode.IsSpace(t) {
			continue
		}

		if t == '(' {
			p.operations.Push(OpenBracket)
			lastTokenIsOp = false
			continue
		}

		if t == ')' {
			if p.operations.Empty() {
				return expression.NewExpression(), fmt.Errorf("неправильные скобки")
			}

			for !p.operations.Empty() && *p.operations.Peek() != OpenBracket {
				err := p.constructNode()
				if err != nil {
					return expression.NewExpression(), err
				}
			}

			p.operations.Pop()
			lastTokenIsOp = false
			continue
		}

		if isOperation(t) {
			op := determineOperation(t)
			if op == expression.Negation {
				p.operations.Push(opToToken[op])
				continue
			}

			if lastTokenIsOp {
				return expression.NewExpression(), fmt.Errorf("некорректный ввод" +
					"(несколько операций следуют друг за другом)")
			}

			lastTokenIsOp = true
			for !p.operations.Empty() && priority[*p.operations.Peek()] > priority[opToToken[op]] {
				err := p.constructNode()
				if err != nil {
					return expression.NewExpression(), err
				}
			}

			p.operations.Push(opToToken[op])
		} else {
			lastTokenIsOp = false
			p.operands.Push(*expression.NewExpressionWithTerm(determineOperand(t)))
		}
	}

	for !p.operations.Empty() {
		err := p.constructNode()
		if err != nil {
			return expression.NewExpression(), err
		}
	}

	return p.operands.Peek(), nil
}

// constructNode создает новый узел из текущих операций и операндов.
func (p *LogicParser) constructNode() error {
	if *p.operations.Peek() == Negation {
		operand := *p.operands.Pop()
		p.operations.Pop()

		operand.Negation(0)
		p.operands.Push(operand)
		return nil
	}

	if p.operands.Len() < 2 || p.operations.Empty() {
		if *p.operations.Peek() == OpenBracket || *p.operations.Peek() == CloseBracket {
			return fmt.Errorf("неправильные скобки")
		}
		return fmt.Errorf("что-то пошло не так при формировании узла")
	}

	// Извлекаем узлы
	rhs := *p.operands.Pop()
	op := *p.operations.Pop()
	lhs := *p.operands.Pop()

	p.operands.Push(expression.Construct(lhs, tokenToOperation[op], rhs))
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
