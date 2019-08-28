package maligned

import (
	"fmt"
	"go/token"
	"go/types"

	malignedAPI "github.com/golangci/maligned"
	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
)

const Name = "maligned"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Tool to detect Go structs that would take less memory if their fields were sorted",
	Run:  run,
}

type Maligned struct {
	SuggestNewOrder bool `mapstructure:"suggest-new"`
}

func run(pass *analysis.Pass) (interface{}, error) {
	maligned := &Maligned{
		SuggestNewOrder: true,
	}
	var createdPkgs []*loader.PackageInfo
	createdPkgs = append(createdPkgs, makeFakeLoaderPackageInfo(pass))
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
	issues := malignedAPI.Run(prog)
	if len(issues) == 0 {
		return nil, nil
	}
	for _, i := range issues {
		text := fmt.Sprintf("[%s] struct of size %d bytes could be of size %d bytes", Name, i.OldSize, i.NewSize)
		if maligned.SuggestNewOrder {
			text += fmt.Sprintf(":\n%s", util.FormatCode(i.NewStructDef))
		}
		pass.Reportf(token.Pos(i.Pos.Offset), text)
	}
	return nil, nil
}

func makeFakeLoaderPackageInfo(pass *analysis.Pass) *loader.PackageInfo {
	var errs []error

	typeInfo := pass.TypesInfo

	return &loader.PackageInfo{
		Pkg:                   pass.Pkg,
		Importable:            true, // not used
		TransitivelyErrorFree: true, // not used

		// use compiled (preprocessed) go files AST;
		// AST linters use not preprocessed go files AST
		Files:  pass.Files,
		Errors: errs,
		Info:   *typeInfo,
	}
}
