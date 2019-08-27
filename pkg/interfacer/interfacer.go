package interfacer

import (
	"mvdan.cc/interfacer/check"

	"go/types"

	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"
)

const Name = "interfacer"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Linter that suggests narrower interface types",
	Run:  run,
}

type Interfacer struct{}

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
	pack := &packages.Package{
		Types:      pass.Pkg,
		Fset:       pass.Fset,
		TypesInfo:  pass.TypesInfo,
		TypesSizes: pass.TypesSizes,
		Syntax:     pass.Files,
	}

	c := new(check.Checker)
	c.Program(prog)
	ssaProg := util.BuildSSAProgram(pack)
	c.ProgramSSA(ssaProg)

	issues, err := c.Check()
	if err != nil {
		return nil, err
	}
	if len(issues) == 0 {
		return nil, nil
	}

	for _, i := range issues {
		pass.Reportf(i.Pos(), i.Message())
	}

	return nil, nil
}
