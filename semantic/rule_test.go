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

package semantic_test

import (
	"os"
	"testing"

	"github.com/ZupIT/horusec-devkit/pkg/enums/confidence"
	"github.com/ZupIT/horusec-devkit/pkg/enums/languages"
	"github.com/ZupIT/horusec-devkit/pkg/enums/severities"
	engine "github.com/ZupIT/horusec-engine"
	"github.com/ZupIT/horusec-engine/semantic"
	"github.com/ZupIT/horusec-engine/semantic/analysis/call"
	"github.com/ZupIT/horusec-engine/semantic/analysis/value"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemanticRule(t *testing.T) {
	src := `
import { spawn } = from 'child_process';

function f(cmd) {
	spawn(cmd);
}
	`

	metadata := engine.Metadata{
		ID:          "HS-TEST-01",
		Name:        t.Name(),
		Severity:    severities.Low.ToString(),
		Confidence:  confidence.Low.ToString(),
		Description: "testing",
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), t.Name())
	require.NoError(t, err, "Expected no error to create temp file: %v", err)

	_, err = tmpFile.WriteString(src)
	require.NoError(t, err, "Expected no error to write on temp file: %v", err)

	rule := semantic.Rule{
		Metadata: metadata,
		Language: languages.Javascript,
		Analyzer: &call.Analyzer{
			Name:      "child_process.spawn",
			ArgsIndex: 1,
			ArgValue:  value.IsConst,
		},
	}

	findings, err := rule.Run(tmpFile.Name())
	require.NoError(t, err, "Expected no error to execute rule: %v", err)

	expectedFindings := []engine.Finding{
		{
			ID:          metadata.ID,
			Name:        metadata.Name,
			Severity:    metadata.Severity,
			CodeSample:  "spawn(cmd)",
			Confidence:  metadata.Confidence,
			Description: metadata.Description,
			SourceLocation: engine.Location{
				Filename: tmpFile.Name(),
				Line:     5,
				Column:   1,
			},
		},
	}

	assert.Equal(t, expectedFindings, findings)
}
