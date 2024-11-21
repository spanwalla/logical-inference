package main

import (
	"fmt"
	"logical-inference/internal/logic"
)

func main() {
	var term = logic.NewTerm(logic.Constant, logic.Negation, 5)
	fmt.Println(term.String())
}
