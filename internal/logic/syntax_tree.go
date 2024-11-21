package logic

import (
	"logical-inference/internal/pkg/alphabet"
	"math"
	"strings"
)

type Value int
type Operation int

const invalidIdx = -1

const (
	Nop Operation = iota
	Negation
	Implication
	Disjunction
	Conjunction
	Xor
	Equivalent
)

var operationNames = map[Operation]string{
	Nop:         "Nop",
	Negation:    "!",
	Implication: "→",
	Disjunction: "*",
	Conjunction: "*",
	Xor:         "+",
	Equivalent:  "=",
}

var oppositeOperations = map[Operation]Operation{
	Nop:         Nop,
	Negation:    Negation,
	Implication: Conjunction,
	Disjunction: Conjunction,
	Conjunction: Implication,
	Xor:         Equivalent,
	Equivalent:  Xor,
}

func (c Operation) String() string {
	if name, ok := operationNames[c]; ok {
		return name
	}
	return "Unknown"
}

func (c Operation) Opposite() Operation {
	if op, ok := oppositeOperations[c]; ok {
		return op
	}
	return Nop
}

func (c Operation) IsCommutative() bool {
	return c != Nop && c != Negation && c != Implication
}

type TermType int

const (
	None TermType = iota
	Constant
	Variable
	Function
)

type Term struct {
	Type TermType
	Op   Operation
	Val  Value
}

func NewTerm(args ...interface{}) Term {
	t := None
	op := Nop
	val := Value(0)

	for _, arg := range args {
		switch v := arg.(type) {
		case TermType:
			t = v
		case Operation:
			op = v
		case Value:
			val = v
		case int:
			val = Value(v)
		}
	}
	return Term{
		Type: t,
		Op:   op,
		Val:  val,
	}
}

func (t Term) String() string {
	var builder strings.Builder

	if t.Type == None {
		builder.WriteString("None")
		return builder.String()
	}

	if t.Type == Function {
		builder.WriteString(t.Op.String())
	} else {
		if t.Op == Negation {
			builder.WriteString(t.Op.String())
		}

		isConst := t.Type == Constant
		letter, err := alphabet.GetLetter(int(math.Abs(float64(t.Val))), !isConst)
		if err != nil {
			return ""
		}
		builder.WriteRune(letter)
	}
	return builder.String()
}

func increaseIdx(idx, offset int) int {
	if idx == invalidIdx {
		return invalidIdx
	}
	return idx + offset
}

func decreaseIdx(idx, offset int) int {
	if idx == invalidIdx || offset > idx {
		return invalidIdx
	}
	return idx - offset
}
