package parser_test

import (
	"fmt"
	"logical-inference/internal/parser"
	"testing"
)

func TestParser_SimpleExpression(t *testing.T) {

	input := "(a|b)*c"
	//expectedNodesCount := 3 // a, b, a+b

	p := parser.NewParser(input)
	nodes, err := p.Parse()

	if err != nil {
		t.Fatalf("Ошибка при разборе выражения: %v", err)
	}

	fmt.Println(nodes)

}
