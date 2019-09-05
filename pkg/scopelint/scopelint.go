package golinters

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/anuvu/bazel-nogo-lint/pkg/result"
	"github.com/anuvu/bazel-nogo-lint/pkg/util"
	"golang.org/x/tools/go/analysis"
)

const Name = "scopelint"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Scopelint checks for unpinned variables in go programs",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	var res []result.Issue

	for _, f := range pass.Files {
		n := Node{
			fset:          pass.Fset,
			dangerObjects: map[*ast.Object]struct{}{},
			unsafeObjects: map[*ast.Object]struct{}{},
			skipFuncs:     map[*ast.FuncLit]struct{}{},
			issues:        &res,
		}
		ast.Walk(&n, f)
	}

	for _, i := range res {
		pass.Reportf(i.IPos, "[%s(%s)] %s", Name, i.FromLinter, i.Text)
	}

	return nil, nil
}

// The code below is copy-pasted from https://github.com/kyoh86/scopelint

// Node represents a Node being linted.
type Node struct {
	fset          *token.FileSet
	dangerObjects map[*ast.Object]struct{}
	unsafeObjects map[*ast.Object]struct{}
	skipFuncs     map[*ast.FuncLit]struct{}
	issues        *[]result.Issue
}

// Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
//nolint:gocyclo,gocritic
func (f *Node) Visit(node ast.Node) ast.Visitor {
	switch typedNode := node.(type) {
	case *ast.ForStmt:
		switch init := typedNode.Init.(type) {
		case *ast.AssignStmt:
			for _, lh := range init.Lhs {
				switch tlh := lh.(type) {
				case *ast.Ident:
					f.unsafeObjects[tlh.Obj] = struct{}{}
				}
			}
		}

	case *ast.RangeStmt:
		// Memory variables declarated in range statement
		switch k := typedNode.Key.(type) {
		case *ast.Ident:
			f.unsafeObjects[k.Obj] = struct{}{}
		}
		switch v := typedNode.Value.(type) {
		case *ast.Ident:
			f.unsafeObjects[v.Obj] = struct{}{}
		}

	case *ast.UnaryExpr:
		if typedNode.Op == token.AND {
			switch ident := typedNode.X.(type) {
			case *ast.Ident:
				if _, unsafe := f.unsafeObjects[ident.Obj]; unsafe {
					f.errorf(ident, "Using a reference for the variable on range scope %s", util.FormatCode(ident.Name))
				}
			}
		}

	case *ast.Ident:
		if _, obj := f.dangerObjects[typedNode.Obj]; obj {
			// It is the naked variable in scope of range statement.
			f.errorf(node, "Using the variable on range scope %s in function literal", util.FormatCode(typedNode.Name))
			break
		}

	case *ast.CallExpr:
		// Ignore func literals that'll be called immediately.
		switch funcLit := typedNode.Fun.(type) {
		case *ast.FuncLit:
			f.skipFuncs[funcLit] = struct{}{}
		}

	case *ast.FuncLit:
		if _, skip := f.skipFuncs[typedNode]; !skip {
			dangers := map[*ast.Object]struct{}{}
			for d := range f.dangerObjects {
				dangers[d] = struct{}{}
			}
			for u := range f.unsafeObjects {
				dangers[u] = struct{}{}
			}
			return &Node{
				fset:          f.fset,
				dangerObjects: dangers,
				unsafeObjects: f.unsafeObjects,
				skipFuncs:     f.skipFuncs,
				issues:        f.issues,
			}
		}
	}
	return f
}

// The variadic arguments may start with link and category types,
// and must end with a format string and any arguments.
// It returns the new Problem.
//nolint:interfacer
func (f *Node) errorf(n ast.Node, format string, args ...interface{}) {
	pos := n.Pos()
	f.errorfAt(pos, format, args...)
}

func (f *Node) errorfAt(pos token.Pos, format string, args ...interface{}) {
	*f.issues = append(*f.issues, result.Issue{
		IPos:       pos,
		Text:       fmt.Sprintf(format, args...),
		FromLinter: Name,
	})
}
