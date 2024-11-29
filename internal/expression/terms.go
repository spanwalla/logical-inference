package expression

import (
	"logical-inference/internal/pkg/alphabet"
	"math"
	"strings"
)

type Value int
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

func NewTerm() Term {
	return Term{
		Type: None,
		Op:   Nop,
		Val:  Value(0),
	}
}

func (t Term) String() string {
	var builder strings.Builder

	if t.Type == None {
		builder.WriteString("None")
	} else if t.Type == Function {
		builder.WriteString(t.Op.String())
	} else {
		if t.Op == Negation {
			builder.WriteString(t.Op.String())
		}

		isConst := t.Type == Constant
		letter, err := alphabet.GetLetter(int(math.Abs(float64(t.Val))), !isConst)
		if err != nil {
			return "Err"
		}
		builder.WriteRune(letter)
	}
	return builder.String()
}
