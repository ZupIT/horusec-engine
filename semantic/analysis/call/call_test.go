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

package call_test

import (
	"testing"

	"github.com/ZupIT/horusec-engine/internal/ir"
	"github.com/ZupIT/horusec-engine/internal/utils/testutil"
	"github.com/ZupIT/horusec-engine/semantic/analysis"
	"github.com/ZupIT/horusec-engine/semantic/analysis/call"
	"github.com/ZupIT/horusec-engine/semantic/analysis/value"
)

type fakeAnalyzerValue struct {
	returnValue bool
}

func (a fakeAnalyzerValue) Run(_ ir.Value) bool { return a.returnValue }

func TestAnalyzerCall(t *testing.T) {
	src := `function f() { eval("console.log('eval')") }`

	testcases := []testutil.TestCaseAnalyzer{
		{
			Name: "MatchArguments",
			Src:  src,
			Analyzer: &call.Analyzer{
				Name:      "eval",
				ArgsIndex: 1,
				ArgValue:  fakeAnalyzerValue{false},
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchArguments",
					StartOffset: 15,
					EndOffset:   42,
					Line:        1,
					Column:      15,
				},
			},
		},
		{
			Name: "NotMatchArguments",
			Src:  src,
			Analyzer: &call.Analyzer{
				Name:      "eval",
				ArgsIndex: 1,
				ArgValue:  fakeAnalyzerValue{true},
			},
			ExpectedIssues: []analysis.Issue{},
		},
		{
			Name: "NotMatchFunctionName",
			Src:  src,
			Analyzer: &call.Analyzer{
				Name:      "fs.readFile",
				ArgsIndex: 1,
				ArgValue:  fakeAnalyzerValue{true},
			},
			ExpectedIssues: []analysis.Issue{},
		},
		{
			Name: "MatchUsingOnlyFunctionName",
			Src:  `function f() { Math.random() }`,
			Analyzer: &call.Analyzer{
				Name:      "Math.random",
				ArgsIndex: call.NoArguments,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchUsingOnlyFunctionName",
					StartOffset: 15,
					EndOffset:   28,
					Line:        1,
					Column:      15,
				},
			},
		},
		{
			Name: "MatchAllArguments",
			Src:  `function f(a, b) { insecureCall("harcoded0", a, b, "harcoded1") }`,
			Analyzer: &call.Analyzer{
				Name:      "insecureCall",
				ArgsIndex: call.AllArguments,
				ArgValue:  value.IsConst,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchAllArguments",
					StartOffset: 19,
					EndOffset:   63,
					Line:        1,
					Column:      19,
				},
			},
		},
		{
			Name: "MatchCallFromVariableValue",
			Src:  `function f(a, b) { const a = insecureCall() }`,
			Analyzer: &call.Analyzer{
				Name:      "insecureCall",
				ArgsIndex: call.NoArguments,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchCallFromVariableValue",
					StartOffset: 29,
					EndOffset:   43,
					Line:        1,
					Column:      29,
				},
			},
		},
		{
			Name: "MatchCallUsingAliases",
			Src: `
import { spawn as exec } from 'child_process';
const process = require('child_process');

function f(cmd) { 
	exec(cmd)
	process.spawn(cmd)
}
			`,
			Analyzer: &call.Analyzer{
				Name:      "child_process.spawn",
				ArgsIndex: call.NoArguments,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchCallUsingAliases",
					StartOffset: 111,
					EndOffset:   120,
					Line:        6,
					Column:      1,
				},
				{
					Filename:    "MatchCallUsingAliases",
					StartOffset: 122,
					EndOffset:   140,
					Line:        7,
					Column:      1,
				},
			},
		},
		{
			Name: "MatchNestedCall",
			Src: `
function f1(cmd) { 
	return Math.Random()
}

function f2(cmd) { 
	const x = 10 * Math.Random()
}
			`,
			Analyzer: &call.Analyzer{
				Name:      "Math.Random",
				ArgsIndex: call.NoArguments,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchNestedCall",
					StartOffset: 29,
					EndOffset:   42,
					Line:        3,
					Column:      8,
				},
				{
					Filename:    "MatchNestedCall",
					StartOffset: 82,
					EndOffset:   95,
					Line:        7,
					Column:      16,
				},
			},
		},
		{
			Name: "MatchCallFromClassMethod",
			Src: `
import fs from 'fs';
class C {
	readFile(path) {
		fs.readFile(path, (data, err) => {
			if (err) {
				console.log(err)
				return
			}
			console.log(data)
		})
	}
}
			`,
			Analyzer: &call.Analyzer{
				Name:      "fs.readFile",
				ArgsIndex: 1,
				ArgValue:  value.IsConst,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchCallFromClassMethod",
					StartOffset: 52,
					EndOffset:   163,
					Line:        5,
					Column:      2,
				},
			},
		},
		{
			Name: "MatchNestedAnonoymounsFunctionCall",
			Src: `
function f() {
	foo(() => {
		insecure();
	})
}
			`,
			Analyzer: &call.Analyzer{
				Name:      "insecure",
				ArgsIndex: call.NoArguments,
			},
			ExpectedIssues: []analysis.Issue{
				{
					Filename:    "MatchNestedAnonoymounsFunctionCall",
					StartOffset: 31,
					EndOffset:   41,
					Line:        4,
					Column:      2,
				},
			},
		},
	}
	testutil.TestAnalayzer(t, testcases)
}
