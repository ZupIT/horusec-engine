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
	Severity       string
	CodeSample     string
	Confidence     string
	Description    string
	SourceLocation Location
}

func RunOutputInJSON(document []Unit, rules []Rule, jsonFilePath string) error {
	report := Run(document, rules)
	bytesToWrite, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	outputFile, err := createOutputFile(jsonFilePath)
	defer func() {
		_ = outputFile.Close()
	}()
	if err != nil {
		return err
	}
	return writeInOutputFile(outputFile, bytesToWrite)
}

func writeInOutputFile(outputFile *os.File, bytesToWrite []byte) error {
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

func createOutputFile(jsonFilePath string) (*os.File, error) {
	if _, err := os.Create(jsonFilePath); err != nil {
		return nil, err
	}
	outputFile, err := os.OpenFile(jsonFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return outputFile, outputFile.Truncate(0)
}

func Run(document []Unit, rules []Rule) (documentFindings []Finding) {
	numberOfUnits := (len(document)) * (len(rules))
	if numberOfUnits == 0 {
		return []Finding{}
	}

	return executeRunByNumberOfUnits(numberOfUnits, document, rules)
}

func executeRunByNumberOfUnits(numberOfUnits int, document []Unit, rules []Rule) (documentFindings []Finding) {
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
		localRule := rule
		if localRule.IsFor(documentUnit.Type()) {
			go func() {
				ruleFindings := documentUnit.Eval(localRule)
				findings <- ruleFindings
			}()
		}
	}
}
