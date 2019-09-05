package gosec

import (
	"fmt"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/anuvu/bazel-nogo-lint/pkg/result"
	"github.com/anuvu/bazel-nogo-lint/pkg/util"
	"github.com/golangci/gosec"
	"github.com/golangci/gosec/rules"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
)

const Name = "gosec"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Inspects source code for security problems",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	gasConfig := gosec.NewConfig()
	enabledRules := rules.Generate()
	logger := log.New(ioutil.Discard, "", 0)
	analyzer := gosec.NewAnalyzer(gasConfig, logger)
	analyzer.LoadRules(enabledRules.Builders())

	var createdPkgs []*loader.PackageInfo
	createdPkgs = append(createdPkgs, util.MakeFakeLoaderPackageInfo(pass))
	allPkgs := map[*types.Package]*loader.PackageInfo{}
	for _, pkg := range createdPkgs {
		pkg := pkg
		allPkgs[pkg.Pkg] = pkg
	}
	prog := &loader.Program{
		Fset:        pass.Fset,
		Imported:    nil,         // not used without .Created in any linter
		Created:     createdPkgs, // all initial packages
		AllPackages: allPkgs,     // all initial packages and their depndencies
	}

	analyzer.ProcessProgram(prog)
	issues, _ := analyzer.Report()
	if len(issues) == 0 {
		return nil, nil
	}

	for _, i := range issues {
		text := fmt.Sprintf("[%s] %s: %s", Name, i.RuleID, i.What) // TODO: use severity and confidence
		var r *result.Range
		line, err := strconv.Atoi(i.Line)
		if err != nil {
			r = &result.Range{}
			if n, rerr := fmt.Sscanf(i.Line, "%d-%d", &r.From, &r.To); rerr != nil || n != 2 {
				//lintCtx.Log.Warnf("Can't convert gosec line number %q of %v to int: %s", i.Line, i, err)
				continue
			}
			line = r.From
		}

		pass.Reportf(token.Pos(line), text)
	}

	return nil, nil
}
