package main

import (
	"fmt"
	"logical-inference/internal/logic"
)

func main() {
	var emptyTerm = logic.NewTerm()
	var term = logic.Term{
		Type: logic.Function,
		Op:   logic.Conjunction,
		Val:  logic.Value(0),
	}
	fmt.Println(emptyTerm.String(), term.String())
}
