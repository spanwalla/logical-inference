package logicparser

import "github.com/spanwalla/logical-inference/internal/expression"

func NewExpressionWithString(expr string) *expression.Expression {
	p := NewLogicParser(expr)
	res, err := p.Parse()
	if err != nil {
		panic(err)
	}
	return res
}
