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

package platforms

import (
	engine "github.com/ZupIT/horusec-engine"
)

// PopulateFindingWithRuleMetadata converts the engine.Metadata field inside the StructuredDataRule to a engine.Finding
// struct
//nolint // change to pointer
func PopulateFindingWithRuleMetadata(
	ruleData StructuredDataRule, filename, codeSample string, line, column int) engine.Finding {
	return engine.Finding{
		ID:          ruleData.ID,
		Name:        ruleData.Name,
		Severity:    ruleData.Severity,
		Confidence:  ruleData.Confidence,
		Description: ruleData.Description,
		CodeSample:  codeSample,
		SourceLocation: engine.Location{
			Filename: filename,
			Line:     line,
			Column:   column,
		},
	}
}
