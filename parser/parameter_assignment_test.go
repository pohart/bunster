package parser_test

import "github.com/yassinebenaid/bunny/ast"

var parameterAssignmentTests = []testCase{
	{`var=value var2='value2'    var3="value3"`, ast.Script{
		ast.ParameterAssignement{
			ast.Assignement{Name: "var", Value: ast.Word("value")},
			ast.Assignement{Name: "var2", Value: ast.Word("value2")},
			ast.Assignement{Name: "var3", Value: ast.Word("value3")},
		},
	}},
	{`var=$var var=${var}`, ast.Script{
		ast.ParameterAssignement{
			ast.Assignement{Name: "var", Value: ast.Var("var")},
			ast.Assignement{Name: "var", Value: ast.Var("var")},
		},
	}},
}
