package prealloc

import (
	"go/ast"
	"go/token"

	"github.com/golangci/prealloc"
	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
)

type PreallocSettings struct {
	Simple     bool
	RangeLoops bool `mapstructure:"range-loops"`
	ForLoops   bool `mapstructure:"range-loops"`
}

const Name = "prealloc"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Finds slice declarations that could potentially be preallocated",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	s := &PreallocSettings{
		Simple:     true,
		RangeLoops: true,
		ForLoops:   false,
	}
	for _, f := range pass.Files {
		hints := prealloc.Check([]*ast.File{f}, s.Simple, s.RangeLoops, s.ForLoops)
		for _, hint := range hints {
			pass.Reportf(token.Pos(hint.Pos), "[%s] Consider preallocating %s", Name, util.FormatCode(hint.DeclaredSliceName))
		}
	}

	return nil, nil
}
