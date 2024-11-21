package logic

type Operation int

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

func (o Operation) String() string {
	if name, ok := operationNames[o]; ok {
		return name
	}
	return "Unknown"
}

func (o Operation) Opposite() Operation {
	if op, ok := oppositeOperations[o]; ok {
		return op
	}
	return Nop
}

func (o Operation) IsCommutative() bool {
	return o != Nop && o != Negation && o != Implication
}
