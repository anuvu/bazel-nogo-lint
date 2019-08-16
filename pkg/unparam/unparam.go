package unparam

import (
	"mvdan.cc/unparam/check"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

type Unparam struct{}

const Name = "depguard"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Reports unused function parameters",
	Run:  run,
}

type UnparamSettings struct {
	CheckExported bool `mapstructure:"check-exported"`
	Algo          string
}

func run(pass *analysis.Pass) (interface{}, error) {
	us := &UnparamSettings{
		Algo: "cha",
	}

	/*if us.Algo != "cha" {
		lintCtx.Log.Warnf("`linters-settings.unparam.algo` isn't supported by the newest `unparam`")
	}*/
	pack := &packages.Package{
		Types:     pass.Pkg,
		Fset:      pass.Fset,
		Syntax:    pass.Files,
		TypesInfo: pass.TypesInfo,
	}
	var packs []*packages.Package
	packs = append(packs, pack)
	c := &check.Checker{}
	c.CheckExportedFuncs(us.CheckExported)
	c.Packages(packs)
	//c.ProgramSSA(lintCtx.SSAProgram)

	unparamIssues, err := c.Check()
	if err != nil {
		return nil, err
	}

	for _, i := range unparamIssues {

		pass.Reportf(i.Pos(), i.Message())
	}

	return nil, nil
}
