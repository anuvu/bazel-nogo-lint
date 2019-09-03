package golinters

import (
	"fmt"
	"go/token"

	duplAPI "github.com/golangci/dupl"
	"github.com/anuvu/bazel-nogo-lint/pkg/config"
	"github.com/anuvu/bazel-nogo-lint/pkg/fsutils"
	"github.com/anuvu/bazel-nogo-lint/pkg/util"
	"golang.org/x/tools/go/analysis"
)

const Name = "dupl"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Tool for code clone detection",
	Run:  run,
}

type Dupl struct{}

func run(pass *analysis.Pass) (interface{}, error) {

	var files []string
	for _, file := range pass.Files {
		files = append(files, file.Name.Name)
	}
	c := config.NewDefault()
	issues, err := duplAPI.Run(files, c.LintersSettings.Dupl.Threshold)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return nil, nil
	}

	for _, i := range issues {
		toFilename, err := fsutils.ShortestRelPath(i.To.Filename(), "")
		if err != nil {
			return nil, err
		}
		dupl := fmt.Sprintf("%s:%d-%d", toFilename, i.To.LineStart(), i.To.LineEnd())
		pass.Reportf(token.Pos(i.From.LineStart()), "[%s] %d-%d lines are duplicate of %s", Name, i.From.LineStart(), i.From.LineEnd(), util.FormatCode(dupl))
	}
	return nil, nil
}
