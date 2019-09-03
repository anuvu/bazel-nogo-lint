package varcheck // nolint:dupl

import (
	"go/token"
	"go/types"

	varcheckAPI "github.com/golangci/check/cmd/varcheck"
	"github.com/anuvu/bazel-nogo-lint/pkg/util"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
)

type Varcheck struct {
	CheckExportedFields bool `mapstructure:"exported-fields"`
}

const Name = "varcheck"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Finds unused global variables and constants",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	varcheck := &Varcheck{
		CheckExportedFields: true,
	}
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
	issues := varcheckAPI.Run(prog, varcheck.CheckExportedFields)
	if len(issues) == 0 {
		return nil, nil
	}

	for _, i := range issues {
		pass.Reportf(token.Pos(i.Pos.Offset), "[%s] %s is unused", Name, util.FormatCode(i.VarName))
	}
	return nil, nil
}
