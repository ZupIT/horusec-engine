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
			Src: `
function f1() { crypto.createHash('sha256') }
function f2() { crypto.createHash('md5') }
`,
			Analyzer: &call.Analyzer{
				Name:      "crypto.createHash",
				ArgsIndex: 1,
				ArgValue: value.Contains{
					Values: []string{"sha256"},
				},
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchContains",
					StartOffset: 63,
					EndOffset:   87,
					Line:        3,
					Column:      16,
				},
			},
		},
		{
			Name: "MatchContainsUsingVariable",
			Src: `
function f1() {
	const algorithmn = 'sha256'
	crypto.createHash(algorithmn) 
}

function f2() {
	const algorithmn = 'md5'
	crypto.createHash(algorithmn) 
}
`,
			Analyzer: &call.Analyzer{
				Name:      "crypto.createHash",
				ArgsIndex: 1,
				ArgValue: value.Contains{
					Values: []string{"sha256"},
				},
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchContainsUsingVariable",
					StartOffset: 124,
					EndOffset:   153,
					Line:        9, Column: 1,
				},
			},
		},
		{
			Name: "MatchContainsUsingGlobalVariable",
			Src: `
const algorithmn = 'md5'
function f() {
	crypto.createHash(algorithmn) 
}
`,
			Analyzer: &call.Analyzer{
				Name:      "crypto.createHash",
				ArgsIndex: 1,
				ArgValue: value.Contains{
					Values: []string{"sha256"},
				},
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchContainsUsingGlobalVariable",
					StartOffset: 42,
					EndOffset:   71,
					Line:        4, Column: 1,
				},
			},
		},
		{
			Name: "MatchPhiValuesIsConst",
			Src: `
function f(arg) {
	if (arg) {
		const x = arg
		const y = 'const true'
	} else {
		const x = 'const'
		const y = 'const false'
	}
	insecureCall(x)
	insecureCall(y)
}
`,
			Analyzer: &call.Analyzer{
				Name:      "insecureCall",
				ArgsIndex: 1,
				ArgValue:  value.IsConst,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchPhiValuesIsConst",
					StartOffset: 132,
					EndOffset:   147,
					Line:        10,
					Column:      1,
				},
			},
		},
	}

	testutil.TestAnalayzer(t, testcases)
}
