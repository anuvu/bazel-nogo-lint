package goconst

import (
	"fmt"
	"go/token"

	goconstAPI "github.com/golangci/goconst"
	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
)

const Name = "goconst"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Finds repeated strings that could be replaced by a constant",
	Run:  run,
}

type Goconst struct {
	MinStringLen        int `mapstructure:"min-len"`
	MinOccurrencesCount int `mapstructure:"min-occurrences"`
}

func run(pass *analysis.Pass) (interface{}, error) {
	var goconstIssues []goconstAPI.Issue
	goconst := &Goconst{
		MinStringLen:        3,
		MinOccurrencesCount: 3,
	}
	cfg := goconstAPI.Config{
		MatchWithConstants: true,
		MinStringLength:    goconst.MinStringLen,
		MinOccurrences:     goconst.MinOccurrencesCount,
	}
	issues, err := goconstAPI.Run(pass.Files, pass.Fset, &cfg)
	if err != nil {
		return nil, err
	}

	goconstIssues = append(goconstIssues, issues...)

	if len(goconstIssues) == 0 {
		return nil, nil
	}

	for _, i := range goconstIssues {
		textBegin := fmt.Sprintf("[%s] string %s has %d occurrences", Name, util.FormatCode(i.Str), i.OccurencesCount)
		var textEnd string
		if i.MatchingConst == "" {
			textEnd = ", make it a constant"
		} else {
			textEnd = fmt.Sprintf(", but such constant %s already exists", util.FormatCode(i.MatchingConst))
		}
		pass.Reportf(token.Pos(i.Pos.Offset), textBegin+textEnd)
	}

	return nil, nil
}
