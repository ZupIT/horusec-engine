package text

import (
	"github.com/ZupIT/horus-engine"
)

type TextUnit struct {
	Files []TextFile
}

func newFinding(ID, filename string, line, column int) engine.Finding {
	return engine.Finding{
		ID: ID,
		SourceLocation: engine.Location{
			Filename: filename,
			Line:     line,
			Column:   column,
		},
	}
}

func (unit TextUnit) evalRegularRule(textRule TextRule, findingsChan chan<- []engine.Finding) {
	for _, file := range unit.Files {
		go func() {
			var findings []engine.Finding

			for _, expression := range textRule.Expressions {
				findingIndexes := expression.FindAllStringIndex(file.Content(), -1)

				if findingIndexes != nil {
					for _, findingIndex := range findingIndexes {
						line, column := file.FindLineAndColumn(findingIndex[0])

						finding := newFinding(
							textRule.ID,
							file.DisplayName,
							line,
							column,
						)

						findings = append(findings, finding)
					}
				}
			}

			findingsChan <- findings
		}()
	}
}

func (unit TextUnit) Type() engine.UnitType {
	return engine.ProgramTextUnit
}

func (unit TextUnit) Eval(rule engine.Rule) (unitFindings []engine.Finding) {
	if len(unit.Files) <= 0 {
		return unitFindings
	}

	chanSize := len(unit.Files) - 1
	findingsChannel := make(chan []engine.Finding, chanSize)

	switch rule.Type() {
	case engine.Regular:
		if textRule, ok := rule.(TextRule); ok {
			go unit.evalRegularRule(textRule, findingsChannel)
		}
	}

	for i := 0; i <= chanSize; i++ {
		fileFindings := <-findingsChannel
		unitFindings = append(unitFindings, fileFindings...)
	}

	close(findingsChannel)

	return unitFindings
}
