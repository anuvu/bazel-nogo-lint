package golint

import (
	"fmt"
	"go/token"

	lintAPI "github.com/golangci/lint-1"
	"golang.org/x/tools/go/analysis"
)

const Name = "golint"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Golint differs from gofmt. Gofmt reformats Go source code, whereas golint prints out style mistakes",
	Run:  run,
}

type Golint struct {
	MinConfidence float64 `mapstructure:"min-confidence"`
}

func run(pass *analysis.Pass) (interface{}, error) {
	golint := &Golint{
		MinConfidence: 0.8,
	}
	l := new(lintAPI.Linter)
	ps, err := l.LintASTFiles(pass.Files, pass.Fset)
	if err != nil {
		return nil, fmt.Errorf("can't lint %d files: %s", len(pass.Files), err)
	}

	if len(ps) == 0 {
		return nil, nil
	}
	for idx := range ps {
		if ps[idx].Confidence >= golint.MinConfidence {
			pass.Reportf(token.Pos(ps[idx].Position.Offset), fmt.Sprintf("[%s] %s", Name, ps[idx].Text))
			// TODO: use p.Link and p.Category
		}
	}
	return nil, nil
}
