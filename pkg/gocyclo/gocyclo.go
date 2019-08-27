package gocyclo

import (
	"go/token"
	"sort"

	gocycloAPI "github.com/golangci/gocyclo/pkg/gocyclo"
	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
)

const MinComplexity = 10
const Name = "gocyclo"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Computes and checks the cyclomatic complexity of functions",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	var stats []gocycloAPI.Stat
	for _, f := range pass.Files {
		stats = gocycloAPI.BuildStats(f, pass.Fset, stats)
	}
	if len(stats) == 0 {
		return nil, nil
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Complexity > stats[j].Complexity
	})

	for _, s := range stats {
		if s.Complexity <= MinComplexity {
			break // Break as the stats is already sorted from greatest to least
		}

		pass.Reportf(token.Pos(s.Pos.Offset), "cyclomatic complexity %d of func %s is high (> %d)", s.Complexity, util.FormatCode(s.FuncName), MinComplexity)
	}

	return nil, nil
}
