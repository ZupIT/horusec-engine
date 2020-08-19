package text

import (
	"regexp"
	"testing"

	"github.com/ZupIT/horus-engine"
)

func TestTextUnitEvalWithRegularMatch(t *testing.T) {
	var exampleGoFile = `package version

import (
	"github.com/ZupIT/horus/development-kit/pkg/utils/logger"
	"github.com/spf13/cobra"
)

type IVersion interface {
	CreateCobraCmd() *cobra.Command
}

type Version struct {
}

func NewVersionCommand() IVersion {
	return &Version{}
}

func (v *Version) CreateCobraCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Actual version installed of the horus",
		Example: "horus version",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.LogPrint(cmd.Short + " is: ")
			return nil
		},
	}
}
`

	var textUnit TextUnit = TextUnit{}
	goTextFile, err := NewTextFile("example/cmd/version.go", []byte(exampleGoFile))

	if err != nil {
		t.Error(err)
	}

	textUnit.Files = append(textUnit.Files, *goTextFile)

	var regularMatchRule TextRule = TextRule{}
	regularMatchRule.Expressions = append(regularMatchRule.Expressions, regexp.MustCompile(`cmd\.Short`))

	rules := []engine.Rule{regularMatchRule}
	program := []engine.Unit{textUnit}

	findings := engine.Run(program, rules)

	if len(findings) < 1 || len(findings) > 1 {
		t.Fatal("Should find only 1 finding")
	}

	for _, finding := range findings {
		t.Logf("Name: %s", finding.SourceLocation.Filename)
		t.Logf("Line: %d", finding.SourceLocation.Line)
		t.Logf("Column: %d", finding.SourceLocation.Column)
	}
}
