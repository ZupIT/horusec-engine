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

package value_test

import (
	"testing"

	"github.com/ZupIT/horusec-engine/internal/utils/testutil"
	"github.com/ZupIT/horusec-engine/semantic/analysis"
	"github.com/ZupIT/horusec-engine/semantic/analysis/call"
	"github.com/ZupIT/horusec-engine/semantic/analysis/value"
)

func TestAnalyzerValue(t *testing.T) {
	testcases := []testutil.TestCaseAnalyzer{
		{
			Name: "MatchIsConstUsingHardocodedConst",
			Src:  `function f() { eval('1+1') }`,
			Analyzer: &call.Analyzer{
				Name:      "eval",
				ArgsIndex: 1,
				ArgValue:  value.IsConst,
			},
			ExpectedIssues: []analysis.Issue{},
		},
		{
			Name: "MatchIsConstUsingConstantVariable",
			Src: `
function f() {
	const s = '1 + 1';
	eval(s) 
}
`,
			Analyzer: &call.Analyzer{
				Name:      "eval",
				ArgsIndex: 1,
				ArgValue:  value.IsConst,
			},
			ExpectedIssues: []analysis.Issue{},
		},
		{
			Name: "MatchIsConstUsingNotConstValue",
			Src:  `function f(s) { eval(s) }`,
			Analyzer: &call.Analyzer{
				Name:      "eval",
				ArgsIndex: 1,
				ArgValue:  value.IsConst,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchIsConstUsingNotConstValue",
					StartOffset: 16,
					EndOffset:   23,
					Line:        1,
					Column:      16,
				},
			},
		},
		{
			Name: "MatchContains",
			Src:  `function f() { crypto.createHash('md5') }`,
			Analyzer: &call.Analyzer{
				Name:      "crypto.createHash",
				ArgsIndex: 1,
				ArgValue: value.Contains{
					Values: []string{"md5"},
				},
			},
			ExpectedIssues: []analysis.Issue{},
		},
		{
			Name: "MatchContainsUsingVariable",
			Src: `
function f() {
	const algorithmn = 'md5'
	crypto.createHash(algorithmn) 
}
`,
			Analyzer: &call.Analyzer{
				Name:      "crypto.createHash",
				ArgsIndex: 1,
				ArgValue: value.Contains{
					Values: []string{"md5"},
				},
			},
			ExpectedIssues: []analysis.Issue{},
		},
	}

	testutil.TestAnalayzer(t, testcases)
}
