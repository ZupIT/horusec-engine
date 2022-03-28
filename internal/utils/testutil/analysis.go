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

package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	javascript "github.com/ZupIT/horusec-engine/internal/horusec-javascript"
	"github.com/ZupIT/horusec-engine/internal/ir"
	"github.com/ZupIT/horusec-engine/semantic/analysis"
)

// TestCaseAnalyzer is a test case used to assert the assertivity of a given Analyzer.
type TestCaseAnalyzer struct {
	Name           string            // Name of testcase.
	Src            string            // Source code that should be used.
	Analyzer       analysis.Analyzer // Analyzer that should be used
	ExpectedIssues []analysis.Issue  // Expected issues that analyzed produces.
}

// TestAnalayzer assert the assertivity of a given Analyzer.
//
// nolint: funlen,gocyclo // There is no need to break this test.
func TestAnalayzer(t *testing.T, testcases []TestCaseAnalyzer) {
	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			ast, err := javascript.ParseFile(tt.Name, []byte(tt.Src))
			require.NoError(t, err, "Expected no error to parse AST: %v", err)

			file := ir.NewFile(ast)
			file.Build()

			issues := make([]analysis.Issue, 0)
			report := func(issue analysis.Issue) {
				issues = append(issues, issue)
			}

			for _, member := range file.Members {
				switch m := member.(type) {
				case *ir.Function:
					run(tt, file, m, report)
					for _, fn := range m.AnonFuncs {
						run(tt, file, fn, report)
					}
				case *ir.Struct:
					for _, method := range m.Methods {
						run(tt, file, method, report)
					}
				}
			}
			assert.Equal(t, tt.ExpectedIssues, issues)
		})
	}
}

func run(tt TestCaseAnalyzer, file *ir.File, fn *ir.Function, report func(analysis.Issue)) {
	tt.Analyzer.Run(&analysis.Pass{
		File:     file,
		Function: fn,
		Report:   report,
	})
}
