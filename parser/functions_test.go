package parser_test

import (
	"github.com/yassinebenaid/bunny/ast"
)

var functionsTests = []testCase{
	{`foo(){ cmd; }`, ast.Script{
		ast.Function{
			Name:    "foo",
			Command: ast.Command{Name: ast.Word("cmd")},
		},
	}},
}
