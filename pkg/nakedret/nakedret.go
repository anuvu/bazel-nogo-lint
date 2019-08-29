package nakedret

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.cisco.com/golinters/pkg/result"
	"golang.org/x/tools/go/analysis"
)

const Name = "nakedret"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Finds naked returns in functions greater than a specified function length",
	Run:  run,
}

type Nakedret struct {
	MaxFuncLines int `mapstructure:"max-func-lines"`
}

type nakedretVisitor struct {
	maxLength int
	f         *token.FileSet
	issues    []result.Issue
}

func (v *nakedretVisitor) processFuncDecl(funcDecl *ast.FuncDecl) {
	file := v.f.File(funcDecl.Pos())
	functionLineLength := file.Position(funcDecl.End()).Line - file.Position(funcDecl.Pos()).Line

	// Scan the body for usage of the named returns
	for _, stmt := range funcDecl.Body.List {
		s, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		if len(s.Results) != 0 {
			continue
		}

		file := v.f.File(s.Pos())
		if file == nil || functionLineLength <= v.maxLength {
			continue
		}
		if funcDecl.Name == nil {
			continue
		}

		v.issues = append(v.issues, result.Issue{
			FromLinter: Name,
			Text: fmt.Sprintf("naked return in func `%s` with %d lines of code",
				funcDecl.Name.Name, functionLineLength),
			Pos: v.f.Position(s.Pos()),
		})
	}
}

func (v *nakedretVisitor) Visit(node ast.Node) ast.Visitor {
	funcDecl, ok := node.(*ast.FuncDecl)
	if !ok {
		return v
	}

	var namedReturns []*ast.Ident

	// We've found a function
	if funcDecl.Type != nil && funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			for _, ident := range field.Names {
				if ident != nil {
					namedReturns = append(namedReturns, ident)
				}
			}
		}
	}

	if len(namedReturns) == 0 || funcDecl.Body == nil {
		return v
	}

	v.processFuncDecl(funcDecl)
	return v
}

func run(pass *analysis.Pass) (interface{}, error) {
	nret := &Nakedret{
		MaxFuncLines: 30,
	}
	for _, f := range pass.Files {
		v := nakedretVisitor{
			maxLength: nret.MaxFuncLines,
			f:         pass.Fset,
		}
		ast.Walk(&v, f)
		for _, issue := range v.issues {
			pass.Reportf(issue.IPos, fmt.Sprintf("[%s] %s", Name, issue.Text))
		}
	}
	return nil, nil
}
