package gochecknoinits

import (
	"go/ast"

	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
)

const Name = "gochecknoinits"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Checks that no init functions are present in Go cod",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := funcDecl.Name.Name
			if name == "init" && funcDecl.Recv.NumFields() == 0 {
				pass.Reportf(funcDecl.Pos(), "[%s] don't use %s function", Name, util.FormatCode(name))
			}
		}
	}

	return nil, nil
}
