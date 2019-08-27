package golinters

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
)

const Name = "gochecknoglobals"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Checks that no globals are present in Go code",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.VAR {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec := spec.(*ast.ValueSpec)
				for _, vn := range valueSpec.Names {
					if isWhitelisted(vn) {
						continue
					}
					pass.Reportf(vn.NamePos, "[%s] %s is a global variable", Name, util.FormatCode(vn.Name))
				}
			}
		}
	}

	return nil, nil
}

func isWhitelisted(i *ast.Ident) bool {
	return i.Name == "_" || looksLikeError(i)
}

// looksLikeError returns true if the AST identifier starts
// with 'err' or 'Err', or false otherwise.
//
// TODO: https://github.com/leighmcculloch/gochecknoglobals/issues/5
func looksLikeError(i *ast.Ident) bool {
	prefix := "err"
	if i.IsExported() {
		prefix = "Err"
	}
	return strings.HasPrefix(i.Name, prefix)
}
