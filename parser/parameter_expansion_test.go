package parser_test

import "github.com/yassinebenaid/bunny/ast"

var parameterExpansionCases = []testCase{
	{`${var}`, ast.Script{Statements: []ast.Statement{
		ast.Command{Name: ast.Var("var")},
	}}},
}