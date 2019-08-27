package depguard

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/loader"

	depguardAPI "github.com/OpenPeeDeeP/depguard"

	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
)

type Depguard struct {
	ListType      string `mapstructure:"list-type"`
	Packages      []string
	IncludeGoRoot bool `mapstructure:"include-go-root"`
}

const Name = "depguard"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Go linter that checks if package imports are in a list of acceptable packages",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Intializing the Linter Config
	var pkgs []string
	pkgs = append(pkgs, pass.Pkg.Name())
	depguard := &Depguard{
		Packages:      pkgs,
		IncludeGoRoot: true,
		ListType:      "",
	}
	dg := &depguardAPI.Depguard{
		Packages:      depguard.Packages,
		IncludeGoRoot: depguard.IncludeGoRoot,
	}
	listType := depguard.ListType
	// Intializing the Loader Config and Program
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
	loadconfig := &loader.Config{
		Cwd:   "",
		Build: nil,
	}
	var found bool
	dg.ListType, found = depguardAPI.StringToListType[strings.ToLower(listType)]
	if !found {
		if listType != "" {
			return nil, fmt.Errorf("unsure what list type %s is", listType)
		}
		dg.ListType = depguardAPI.LTBlacklist
	}
	issues, err := dg.Run(loadconfig, prog)
	if err != nil {
		return nil, err
	}
	if len(issues) == 0 {
		return nil, nil
	}
	msgSuffix := "is in the blacklist"
	if dg.ListType == depguardAPI.LTWhitelist {
		msgSuffix = "is not in the whitelist"
	}
	for _, i := range issues {
		pass.Reportf(token.Pos(i.Position.Offset), "%s %s", util.FormatCode(i.PackageName), msgSuffix)
	}
	return nil, nil
}
