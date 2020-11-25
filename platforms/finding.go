package platforms

import (
	engine "github.com/ZupIT/horusec-engine"
)

// PopulateFindingWithRuleMetadata converts the engine.Metadata field inside the StructuredDataRule to a engine.Finding
// struct
func PopulateFindingWithRuleMetadata(ruleData StructuredDataRule, filename, codeSample string, line, column int) engine.Finding {
	return engine.Finding{
		ID:          ruleData.ID,
		Name:        ruleData.Name,
		Severity:    ruleData.Severity,
		Confidence:  ruleData.Confidence,
		Description: ruleData.Description,

		CodeSample: codeSample,

		SourceLocation: engine.Location{
			Filename: filename,
			Line:     line,
			Column:   column,
		},
	}
}
