package engine

import (
	"encoding/json"
	"fmt"
	"os"
)

type Unit interface {
	Type() UnitType
	Eval(Rule) []Finding
}

type Finding struct {
	ID             string
	Name           string
	Description    string
	SourceLocation Location
}

type Report struct {
	ID             string   // Comes from Advisory::GetID/0
	Name           string   // Comes from Advisory::GetName/0
	Description    string   // Comes from Advisory::GetDescription/0
	SourceLocation Location // Comes from the Finding
}

func RunAndBuildReportInFile(document []Unit, rules []Rule, advisories []Advisory, outputFilePath string) error {
	report := RunAndBuildReport(document, rules, advisories)
	bytesToWrite, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if _, err := os.Create(outputFilePath); err != nil {
		return err
	}
	outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	if err = outputFile.Truncate(0); err != nil {
		return err
	}
	bytesWritten, err := outputFile.Write(bytesToWrite)
	if err != nil {
		return err
	}
	if bytesWritten != len(bytesToWrite) {
		return fmt.Errorf("bytes written and length of bytes to write is not equal: %v", map[string]interface{}{
			"bytesWritten": bytesWritten,
			"bytesToWrite": string(bytesToWrite),
		})
	}
	return nil
}

func RunAndBuildReport(document []Unit, rules []Rule, advisories []Advisory) (programReport []Report) {
	findings := Run(document, rules)
	for _, advisory := range advisories {
		for _, finding := range findings {
			if finding.ID == advisory.GetID() {
				report := Report{
					ID:             advisory.GetID(),
					Name:           advisory.GetName(),
					Description:    advisory.GetDescription(),
					SourceLocation: finding.SourceLocation,
				}
				programReport = append(programReport, report)
			}
		}
	}
	if programReport == nil {
		return []Report{}
	}
	return programReport
}

func Run(document []Unit, rules []Rule) (documentFindings []Finding) {
	if len(document) < 1 || len(rules) < 1 {
		return []Finding{}
	}

	numberOfUnits := (len(document)) * (len(rules))

	documentFindingsChannel := make(chan []Finding, numberOfUnits)

	for _, documentUnit := range document {
		localDocumentUnit := documentUnit
		go execRulesInDocumentUnit(rules, localDocumentUnit, documentFindingsChannel)
	}

	for i := 1; i <= numberOfUnits; i++ {
		unitFindings := <-documentFindingsChannel
		documentFindings = append(documentFindings, unitFindings...)
	}

	close(documentFindingsChannel)

	return documentFindings
}

func execRulesInDocumentUnit(rules []Rule, documentUnit Unit, findings chan<- []Finding) {
	for _, rule := range rules {
		if rule.IsFor(documentUnit.Type()) {
			localRule := rule
			go func() {
				ruleFindings := documentUnit.Eval(localRule)
				findings <- ruleFindings
			}()
		}
	}
}
