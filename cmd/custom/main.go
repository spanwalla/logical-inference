package main

import (
	"fmt"
	"logical-inference/internal/parser"
)

func main() {
	exprParser := parser.NewParser("(a>b)|c")
	expression, err := exprParser.Parse()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(expression.Size())
	fmt.Println(expression)
}
