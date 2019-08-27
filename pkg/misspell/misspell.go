package misspell

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/golangci/misspell"
	"golang.cisco.com/golinters/pkg/util"
	"golang.org/x/tools/go/analysis"
)

const Name = "misspell"

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  "Checks the spelling error in code",
	Run:  run,
}

type Misspell struct {
	Locale      string
	IgnoreWords []string `mapstructure:"ignore-words"`
}

func run(pass *analysis.Pass) (interface{}, error) {
	r := misspell.Replacer{
		Replacements: misspell.DictMain,
	}

	// Figure out regional variations
	settings := &Misspell{
		Locale: "US",
	}
	locale := settings.Locale
	switch strings.ToUpper(locale) {
	case "":
		// nothing
	case "US":
		r.AddRuleList(misspell.DictAmerican)
	case "UK", "GB":
		r.AddRuleList(misspell.DictBritish)
	case "NZ", "AU", "CA":
		return nil, fmt.Errorf("unknown locale: %q", locale)
	}

	if len(settings.IgnoreWords) != 0 {
		r.RemoveRule(settings.IgnoreWords)
	}

	r.Compile()
	var files []string
	for _, file := range pass.Files {
		files = append(files, file.Name.Name)
	}
	for _, f := range files {
		err := runOnFile(f, &r, pass)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func runOnFile(fileName string, r *misspell.Replacer, pass *analysis.Pass) error {
	fileContent, err := util.GetFileBytes(fileName)
	if err != nil {
		return fmt.Errorf("can't get file %s contents: %s", fileName, err)
	}

	// use r.Replace, not r.ReplaceGo because r.ReplaceGo doesn't find
	// issues inside strings: it searches only inside comments. r.Replace
	// searches all words: it treats input as a plain text. A standalone misspell
	// tool uses r.Replace by default.
	_, diffs := r.Replace(string(fileContent))
	for _, diff := range diffs {
		text := fmt.Sprintf("`%s` is a misspelling of `%s`", diff.Original, diff.Corrected)
		pass.Reportf(token.Pos(diff.Line), text)
	}
	return nil
}
