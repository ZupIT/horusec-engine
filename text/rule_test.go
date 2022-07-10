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

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name             string
		filepath         string
		matchType        MatchType
		filter           string
		expressions      []*regexp.Regexp
		expectedFindings int
	}{
		{
			name:      "Should return 1 findings with match type AndMatch and go sample",
			filepath:  filepath.Join("examples", "go", "example1", "api", "server.go"),
			matchType: AndMatch,
			filter:    "**/*.go",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`Logger\.Fatal`),
				regexp.MustCompile(`echoInstance.Start`),
			},
			expectedFindings: 1,
		},
		{
			name:      "Should return 3 findings with match type OrMatch and go sample",
			filepath:  filepath.Join("examples", "go", "example1", "api", "server.go"),
			matchType: OrMatch,
			filter:    "**/*.go",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`echoInstance\.Use\(middleware`),
			},
			expectedFindings: 3,
		},
		{
			name:      "Should return 1 findings with match type NotMatch and go sample",
			filepath:  filepath.Join("examples", "go", "example1", "api", "server.go"),
			matchType: NotMatch,
			filter:    "**/*.go",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`should-not-match`),
			},
			expectedFindings: 1,
		},
		{
			name:      "Should return 0 findings with match type AndMatch and go sample",
			filepath:  filepath.Join("examples", "go", "example1", "api", "server.go"),
			matchType: AndMatch,
			filter:    "**/*.go",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`should-not-match`),
				regexp.MustCompile(`echoInstance.Start`),
			},
			expectedFindings: 0,
		},
		{
			name:      "Should return 0 findings with match type OrMatch and go sample",
			filepath:  filepath.Join("examples", "go", "example1", "api", "server.go"),
			matchType: OrMatch,
			filter:    "**/*.go",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`should-not-match`),
			},
			expectedFindings: 0,
		},
		{
			name:      "Should return 0 findings with match type NotMatch and go sample",
			filepath:  filepath.Join("examples", "go", "example1", "api", "server.go"),
			matchType: NotMatch,
			filter:    "**/*.go",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`Logger\.Fatal`),
			},
			expectedFindings: 0,
		},
		{
			name:      "Should return 1 findings with match type AndMatch and python sample",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: AndMatch,
			filter:    "**/*.py",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`secret =`),
				regexp.MustCompile(`password =`),
				regexp.MustCompile(`command =`),
			},
			expectedFindings: 1,
		},
		{
			name:      "Should return 3 findings with match type OrMatch and python sample",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: OrMatch,
			filter:    "**/*.py",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`secret =`),
				regexp.MustCompile(`password =`),
				regexp.MustCompile(`command =`),
				regexp.MustCompile(`print(secret)`),
			},
			expectedFindings: 3,
		},
		{
			name:      "Should return 1 findings with match type NotMatch and python sample",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: NotMatch,
			filter:    "**/*.py",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`should-not-match`),
			},
			expectedFindings: 1,
		},
		{
			name:      "Should return 0 findings with match type AndMatch and python sample",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: AndMatch,
			filter:    "**/*.py",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`secret =`),
				regexp.MustCompile(`password =`),
				regexp.MustCompile(`should-not-match`),
			},
			expectedFindings: 0,
		},
		{
			name:      "Should return 0 findings with match type OrMatch and python sample",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: OrMatch,
			filter:    "**/*.py",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`should-not-match`),
				regexp.MustCompile(`should-not-match`),
			},
			expectedFindings: 0,
		},
		{
			name:      "Should return 0 findings with match type NotMatch and python sample",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: NotMatch,
			filter:    "**/*.py",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`secret =`),
			},
			expectedFindings: 0,
		},
		{
			name:      "Should return 0 findings with non existing filename",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: NotMatch,
			filter:    "**/non-existing-file.py",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`secret =`),
			},
			expectedFindings: 0,
		},
		{
			name:      "Should not return 3 findings with non existing filtered filename",
			filepath:  filepath.Join("examples", "python", "example1", "main.py"),
			matchType: OrMatch,
			filter:    "**/*.go",
			expressions: []*regexp.Regexp{
				regexp.MustCompile(`secret =`),
				regexp.MustCompile(`password =`),
				regexp.MustCompile(`command =`),
				regexp.MustCompile(`print(secret)`),
			},
			expectedFindings: 0,
		},
	}

	for index, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rule := &Rule{
				Type:        testCase.matchType,
				Expressions: testCase.expressions,
			}
			rule.Metadata.Filter = testCase.filter

			findings, err := rule.Run(testCase.filepath)
			assert.NoError(t, err)
			assert.Lenf(
				t,
				findings,
				testCase.expectedFindings,
				"should contains %d findings, but has %d. Test case %d",
				testCase.expectedFindings, len(findings), index,
			)
		})
	}
}
