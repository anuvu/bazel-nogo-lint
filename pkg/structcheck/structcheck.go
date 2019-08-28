package structcheck // nolint:dupl

import (
	"go/token"
	"go/types"

	structcheckAPI "github.com/golangci/check/cmd/structcheck"
	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
)

const Name = "structcheck"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Finds unused struct fields",
	Run:  run,
}

type Structcheck struct {
	CheckExportedFields bool `mapstructure:"exported-fields"`
}

func run(pass *analysis.Pass) (interface{}, error) {
	structcheck := &Structcheck{
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
	issues := structcheckAPI.Run(prog, structcheck.CheckExportedFields)
	if len(issues) == 0 {
		return nil, nil
	}

	for _, i := range issues {
		pass.Reportf(token.Pos(i.Pos.Offset), "[%s] %s is unused", Name, util.FormatCode(i.FieldName))
	}

	return nil, nil
}
