// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package text

import (
	"path/filepath"
	"regexp"
	"testing"

	engine "github.com/ZupIT/horusec-engine"
)

func TestTextUnitEvalWithRegularMatch(t *testing.T) {
	var exampleGoFile = `package version

import (
	"github.com/ZupIT/horusec/development-kit/pkg/utils/logger"
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
		Short:   "Actual version installed of the horusec",
		Example: "horusec version",
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
	"github.com/ZupIT/horusec/development-kit/pkg/utils/logger"
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
		Short:   "Actual version installed of the horusec",
		Example: "horusec version",
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

func TestTextunitEvalWithRegularMatchWithMultipleRules(t *testing.T) {
	javaFileContent := `package com.mycompany.app;

import java.util.Random;

/**
 * Hello world!
 *
 */
public class App 
{
    public static void main( String[] args )
    {
        String password = "Ch@ng3m3"
        Random rand = new Random();
        System.out.println(rand.nextInt(50));
        System.out.println( "Hello World!" );
        System.out.println( "Actual password" + password );
    }
}`

	var textUnit TextUnit = TextUnit{}

	javaFile, err := NewTextFile("example/src/main.java", []byte(javaFileContent))

	if err != nil {
		t.Fatal(err)
	}

	textUnit.Files = append(textUnit.Files, javaFile)

	var regularMatchRule TextRule = TextRule{}
	regularMatchRule.Type = Regular
	regularMatchRule.Description = "Finds java.util.Random imports"
	regularMatchRule.Expressions = append(regularMatchRule.Expressions, regexp.MustCompile(`java\.util\.Random`))

	var anotherRegularMatchRule TextRule = TextRule{}
	anotherRegularMatchRule.Type = Regular
	anotherRegularMatchRule.Description = "Finds hardcoded passwords"
	anotherRegularMatchRule.Expressions = append(anotherRegularMatchRule.Expressions, regexp.MustCompile(`(password\s*=\s*['|\"]\w+[[:print:]]*['|\"])|(pass\s*=\s*['|\"]\w+['|\"]\s)|(pwd\s*=\s*['|\"]\w+['|\"]\s)|(passwd\s*=\s*['|\"]\w+['|\"]\s)|(senha\s*=\s*['|\"]\w+['|\"])`))

	rules := []engine.Rule{regularMatchRule, anotherRegularMatchRule}
	program := []engine.Unit{textUnit}

	findings := engine.Run(program, rules)

	for _, finding := range findings {
		t.Log(finding.Description)
		t.Log(finding.SourceLocation)
	}

	if len(findings) < 2 || len(findings) > 2 {
		t.Fatalf("Should find only 2 finding, but found %d", len(findings))
	}

}

func TestTextunitEvalWithAndMatch(t *testing.T) {
	javaFileContent := `package com.mycompany.app;

import java.util.Random;

/**
 * Hello world!
 *
 */
public class App 
{
    public static void main( String[] args )
    {
        String password = "Ch@ng3m3"
        Random rand = new Random();
        System.out.println(rand.nextInt(50));
        System.out.println( "Hello World!" );
        System.out.println( "Actual password" + password );
    }
}`

	var textUnit TextUnit = TextUnit{}

	javaFile, err := NewTextFile("example/src/main.java", []byte(javaFileContent))

	if err != nil {
		t.Fatal(err)
	}

	textUnit.Files = append(textUnit.Files, javaFile)

	var andMatchRule TextRule = TextRule{}
	andMatchRule.Description = "Finds java.util.Random imports"
	andMatchRule.Type = AndMatch
	andMatchRule.Expressions = append(andMatchRule.Expressions, regexp.MustCompile(`java\.util\.Random`))
	andMatchRule.Expressions = append(andMatchRule.Expressions, regexp.MustCompile(`rand\.\w+\(`))

	rules := []engine.Rule{andMatchRule}
	program := []engine.Unit{textUnit}

	findings := engine.Run(program, rules)

	for _, finding := range findings {
		t.Log(finding.Description)
		t.Log(finding.SourceLocation)
	}

	if len(findings) != 1 {
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
	benchFiles := []string{
		"benchmark.perf",
		"benchmark1.perf",
		"benchmark2.perf",
		"benchmark3.perf",
		"benchmark4.perf",
	}

	var textUnit TextUnit = TextUnit{}

	var summaryIdentifier TextRule = TextRule{}
	summaryIdentifier.Expressions = append(summaryIdentifier.Expressions, regexp.MustCompile(`Summary`))

	var instanceIdentifier TextRule = TextRule{}
	instanceIdentifier.Expressions = append(instanceIdentifier.Expressions, regexp.MustCompile(`Instance`))

	var staticMethodsIdentifier TextRule = TextRule{}
	staticMethodsIdentifier.Expressions = append(staticMethodsIdentifier.Expressions, regexp.MustCompile(`static`))

	rules := []engine.Rule{summaryIdentifier, instanceIdentifier, staticMethodsIdentifier}

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

func BenchmarkHeavyGolangWithMultipleUnits(b *testing.B) {
	benchFiles := []string{
		"benchmark.perf",
		"benchmark1.perf",
		"benchmark2.perf",
		"benchmark3.perf",
		"benchmark4.perf",
	}

	var summaryIdentifier TextRule = TextRule{}
	summaryIdentifier.Expressions = append(summaryIdentifier.Expressions, regexp.MustCompile(`Summary`))

	var instanceIdentifier TextRule = TextRule{}
	instanceIdentifier.Expressions = append(instanceIdentifier.Expressions, regexp.MustCompile(`Instance`))

	var staticMethodsIdentifier TextRule = TextRule{}
	staticMethodsIdentifier.Expressions = append(staticMethodsIdentifier.Expressions, regexp.MustCompile(`static`))

	rules := []engine.Rule{summaryIdentifier, instanceIdentifier, staticMethodsIdentifier}

	program := []engine.Unit{}

	for _, benchFileName := range benchFiles {
		var textUnit TextUnit = TextUnit{}
		benchFile, err := ReadAndCreateTextFile(filepath.Join("samples", benchFileName))

		if err != nil {
			b.Fatal(err)
		}

		textUnit.Files = append(textUnit.Files, benchFile)
		for i := 0; i <= 1024; i++ {
			program = append(program, textUnit)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.Run(program, rules)
	}
}
