package text

import (
	"path/filepath"
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

	textUnit.Files = append(textUnit.Files, goTextFile)

	var regularMatchRule TextRule = TextRule{}
	regularMatchRule.Type = Regular
	regularMatchRule.Expressions = append(regularMatchRule.Expressions, regexp.MustCompile(`cmd\.Short`))

	rules := []engine.Rule{regularMatchRule}
	program := []engine.Unit{textUnit}

	findings := engine.Run(program, rules)

	for _, finding := range findings {
		t.Log(finding.SourceLocation)
	}

	if len(findings) < 1 || len(findings) > 1 {
		t.Fatal("Should find only 1 finding")
	}

}

func TestTextUnitEvalWithRegularMatchWithNoPositiveMatches(t *testing.T) {
	var exampleGoFile = `package version

type Version struct {
}

`

	var textUnit TextUnit = TextUnit{}
	goTextFile, err := NewTextFile("example/cmd/version.go", []byte(exampleGoFile))

	if err != nil {
		t.Error(err)
	}

	textUnit.Files = append(textUnit.Files, goTextFile)

	var regularMatchRule TextRule = TextRule{}
	regularMatchRule.Type = Regular
	regularMatchRule.Expressions = append(regularMatchRule.Expressions, regexp.MustCompile(`cmd\.Short`))

	rules := []engine.Rule{regularMatchRule}
	program := []engine.Unit{textUnit}

	findings := engine.Run(program, rules)

	for _, finding := range findings {
		t.Log(finding.SourceLocation)
	}

	if len(findings) > 0 {
		t.Fatal("Should not find anything")
	}
}

func TestTextUnitEvalWithRegularMatchWithMultipleFiles(t *testing.T) {
	var examplePositiveGoFile = `package version

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

	var exampleNegativeGoFile = `package version

type Version struct {
}

`

	var textUnit TextUnit = TextUnit{}
	goPositiveTextFile, err := NewTextFile("example/cmd/version.go", []byte(examplePositiveGoFile))

	if err != nil {
		t.Error(err)
	}

	goNegativeTextFile, err := NewTextFile("example/cmd/struct.go", []byte(exampleNegativeGoFile))

	if err != nil {
		t.Error(err)
	}

	textUnit.Files = append(textUnit.Files, goPositiveTextFile)
	textUnit.Files = append(textUnit.Files, goNegativeTextFile)

	var regularMatchRule TextRule = TextRule{}
	regularMatchRule.Type = Regular
	regularMatchRule.Expressions = append(regularMatchRule.Expressions, regexp.MustCompile(`cmd\.Short`))

	rules := []engine.Rule{regularMatchRule}
	program := []engine.Unit{textUnit}

	findings := engine.Run(program, rules)

	for _, finding := range findings {
		t.Log(finding.SourceLocation)
	}

	if len(findings) < 1 || len(findings) > 1 {
		t.Fatalf("Should find only 1 finding, but found %d", len(findings))
	}
}

/*
 *
 *
 * ******* Benchmarks ********
 *
 */

func BenchmarkHeavyGolangWithSingleTextUnit(b *testing.B) {
	benchFiles := []string{"benchmark.perf.go", "benchmark1.perf.go", "benchmark2.perf.go"}
	var textUnit TextUnit = TextUnit{}

	var summaryIdentifier TextRule = TextRule{}
	summaryIdentifier.Expressions = append(summaryIdentifier.Expressions, regexp.MustCompile(`Summary`))

	var instanceIdentifier TextRule = TextRule{}
	instanceIdentifier.Expressions = append(instanceIdentifier.Expressions, regexp.MustCompile(`Instance`))

	rules := []engine.Rule{summaryIdentifier, instanceIdentifier}

	for _, benchFileName := range benchFiles {
		benchFile, err := ReadAndCreateTextFile(filepath.Join("samples", benchFileName))

		if err != nil {
			b.Fatal(err)
		}

		textUnit.Files = append(textUnit.Files, benchFile)
	}

	program := []engine.Unit{textUnit}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.Run(program, rules)
	}
}
