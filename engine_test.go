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

package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ruleMock struct {
	findings []Finding
	err      error
}

func newRuleMock(findings []Finding, err error) *ruleMock {
	return &ruleMock{
		findings: findings,
		err:      err,
	}
}

// Run will return a total of findings depending on total of file paths found in informed project path and total of
// findings passed to the mock (ruleMock.findings * file paths)
func (r *ruleMock) Run(_ string) ([]Finding, error) {
	return r.findings, r.err
}

func TestEngineRun(t *testing.T) {
	testcases := []struct {
		name             string
		projectPath      string
		extensions       []string
		rules            Rule
		err              bool
		expectedFindings int
	}{
		{
			name:             "Should run without errors and return 225 findings",
			projectPath:      filepath.Join("text", "examples"),
			extensions:       []string{AcceptAnyExtension},
			rules:            newRuleMock([]Finding{{}}, nil),
			expectedFindings: 237,
			err:              false,
		},
		{
			name:             "Should run without errors and return 24 findings",
			projectPath:      filepath.Join("text", "examples"),
			extensions:       []string{".ex"},
			rules:            newRuleMock([]Finding{{}}, nil),
			expectedFindings: 24,
			err:              false,
		},
		{
			name:             "Should run without errors and return 3 findings",
			projectPath:      filepath.Join("text", "examples"),
			extensions:       []string{".py"},
			rules:            newRuleMock([]Finding{{}}, nil),
			expectedFindings: 3,
			err:              false,
		},
		{
			name:             "Should run without errors and return 4 findings",
			projectPath:      filepath.Join("text", "examples"),
			extensions:       []string{".go"},
			rules:            newRuleMock([]Finding{{}}, nil),
			expectedFindings: 4,
			err:              false,
		},
		{
			name:             "Should run without errors and return 0 findings",
			projectPath:      filepath.Join("text", "examples"),
			extensions:       []string{".invalidExt"},
			rules:            newRuleMock([]Finding{{}}, nil),
			expectedFindings: 0,
			err:              false,
		},
		{
			name:             "Should return error when invalid project path",
			projectPath:      "invalidPath",
			extensions:       []string{AcceptAnyExtension},
			rules:            newRuleMock(nil, nil),
			expectedFindings: 0,
			err:              true,
		},
		{
			name:             "Should return error when failed to run rule",
			extensions:       []string{AcceptAnyExtension},
			expectedFindings: 0,
			rules:            newRuleMock(nil, errors.New("test error")),
			projectPath:      filepath.Join("text", "examples"),
			err:              true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			engine := NewEngine(0, testcase.extensions...)

			findings, err := engine.Run(context.Background(), testcase.projectPath, testcase.rules)
			if testcase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, findings, testcase.expectedFindings)
		})
	}
}
