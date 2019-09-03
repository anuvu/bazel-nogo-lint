package deadcode

import (
	"go/token"
	"go/types"

	deadcodeAPI "github.com/golangci/go-misc/deadcode"
	"github.com/anuvu/bazel-nogo-lint/pkg/util"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
)

const Name = "deadcode"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Finds unused code",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
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
	issues, err := deadcodeAPI.Run(prog)
	if err != nil {
		return nil, err
	}
	for _, i := range issues {
		pass.Reportf(token.Pos(i.Pos.Offset), "[%s] %s is unused", Name, util.FormatCode(i.UnusedIdentName))
	}
	return nil, nil
}
