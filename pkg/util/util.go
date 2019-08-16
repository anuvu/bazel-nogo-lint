package util

import (
	"fmt"
	"go/ast"
	"io/ioutil"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
)

func MakeFakeLoaderPackageInfo(pass *analysis.Pass) *loader.PackageInfo {
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

func FormatCode(code string) string {
	if strings.Contains(code, "`") {
		return code // TODO: properly escape or remove
	}

	return fmt.Sprintf("`%s`", code)
}
func GetFileBytes(filePath string) ([]byte, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
func BuildSSAProgram(pkg *packages.Package) *ssa.Program {
	var pkgs []*packages.Package
	pkgs = append(pkgs, pkg)
	ssaProg, _ := ssautil.Packages(pkgs, ssa.GlobalDebug)
	ssaProg.Build()
	return ssaProg
}
func GetAllFileNames(files []*ast.File) []string {
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name.Name)
	}
	return filenames
}
