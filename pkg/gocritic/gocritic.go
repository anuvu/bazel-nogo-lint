package gocritic

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	"github.com/anuvu/bazel-nogo-lint/pkg/config"
	"github.com/anuvu/bazel-nogo-lint/pkg/result"
	"github.com/anuvu/bazel-nogo-lint/pkg/util"
	"github.com/go-lintpack/lintpack"
	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
)

const Name = "gocritic"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Computes and checks the cyclomatic complexity of functions",
	Run:  run,
}

type Gocritic struct{}

func (Gocritic) Name() string {
	return "gocritic"
}

func (Gocritic) Desc() string {
	return "The most opinionated Go source code linter"
}

func normalizeCheckerInfoParams(info *lintpack.CheckerInfo) lintpack.CheckerParams {
	// lowercase info param keys here because golangci-lint's config parser lowercases all strings
	ret := lintpack.CheckerParams{}
	for k, v := range info.Params {
		ret[strings.ToLower(k)] = v
	}

	return ret
}

func configureCheckerInfo(info *lintpack.CheckerInfo, allParams map[string]config.GocriticCheckSettings) error {
	params := allParams[strings.ToLower(info.Name)]
	if params == nil { // no config for this checker
		return nil
	}

	infoParams := normalizeCheckerInfoParams(info)
	for k, p := range params {
		v, ok := infoParams[k]
		if ok {
			v.Value = p
			continue
		}

		// param `k` isn't supported
		if len(info.Params) == 0 {
			return fmt.Errorf("checker %s config param %s doesn't exist: checker doesn't have params",
				info.Name, k)
		}

		var supportedKeys []string
		for sk := range info.Params {
			supportedKeys = append(supportedKeys, sk)
		}
		sort.Strings(supportedKeys)

		return fmt.Errorf("checker %s config param %s doesn't exist, all existing: %s",
			info.Name, k, supportedKeys)
	}

	return nil
}

func buildEnabledCheckers(lintpackCtx *lintpack.Context) ([]*lintpack.Checker, error) {
	s := &config.GocriticSettings{}
	allParams := s.GetLowercasedParams()

	var enabledCheckers []*lintpack.Checker
	for _, info := range lintpack.GetCheckersInfo() {
		if !s.IsCheckEnabled(info.Name) {
			continue
		}

		if err := configureCheckerInfo(info, allParams); err != nil {
			return nil, err
		}

		c := lintpack.NewChecker(lintpackCtx, info)
		enabledCheckers = append(enabledCheckers, c)
	}

	return enabledCheckers, nil
}

func run(pass *analysis.Pass) (interface{}, error) {
	sizes := types.SizesFor("gc", runtime.GOARCH)
	lintpackCtx := lintpack.NewContext(pass.Fset, sizes)

	enabledCheckers, err := buildEnabledCheckers(lintpackCtx)
	if err != nil {
		return nil, err
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

	issuesCh := make(chan result.Issue, 1024)
	var panicErr error
	go func() {
		defer func() {
			if err := recover(); err != nil {
				panicErr = fmt.Errorf("panic occurred: %s", err)
				log.Warnf("Panic: %s", string(debug.Stack()))
			}
		}()

		for _, pkgInfo := range prog.InitialPackages() {
			lintpackCtx.SetPackageInfo(&pkgInfo.Info, pkgInfo.Pkg)
			runOnPackage(lintpackCtx, enabledCheckers, pkgInfo, issuesCh)
		}
		close(issuesCh)
	}()

	var res []result.Issue
	for i := range issuesCh {
		res = append(res, i)
	}
	if panicErr != nil {
		return nil, panicErr
	}

	for _, i := range res {
		pass.Reportf(i.IPos, "[%s(%s)] %s", Name, i.FromLinter, i.Text)
	}

	return nil, nil
}

func runOnPackage(lintpackCtx *lintpack.Context, checkers []*lintpack.Checker,
	pkgInfo *loader.PackageInfo, ret chan<- result.Issue) {

	for _, f := range pkgInfo.Files {
		filename := filepath.Base(lintpackCtx.FileSet.Position(f.Pos()).Filename)
		lintpackCtx.SetFileInfo(filename, f)

		runOnFile(lintpackCtx, f, checkers, ret)
	}
}

func runOnFile(ctx *lintpack.Context, f *ast.File, checkers []*lintpack.Checker,
	ret chan<- result.Issue) {

	var wg sync.WaitGroup
	wg.Add(len(checkers))
	for _, c := range checkers {
		// All checkers are expected to use *lint.Context
		// as read-only structure, so no copying is required.
		go func(c *lintpack.Checker) {
			defer wg.Done()

			for _, warn := range c.Check(f) {
				pos := ctx.FileSet.Position(warn.Node.Pos())
				ret <- result.Issue{
					Pos:        pos,
					Text:       fmt.Sprintf("%s: %s", c.Info.Name, warn.Text),
					FromLinter: Name,
				}
			}
		}(c)
	}

	wg.Wait()
}
