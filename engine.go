package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/ZupIT/horusec/development-kit/pkg/utils/logger"
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

func Run(document []Unit, rules []Rule) (documentFindings []Finding) {
	numberOfUnits := (len(document)) * (len(rules))
	if numberOfUnits == 0 {
		return []Finding{}
	}

	return executeRunByNumberOfUnits(numberOfUnits, document, rules)
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

func RunMaxUnitsByAnalysis(document []Unit, rules []Rule, maxUnitPerAnalysis int) (documentFindings []Finding) {
	listDocuments := breakTextUnitsIntoLimitOfUnit(document, maxUnitPerAnalysis)
	for key, units := range listDocuments {
		logger.LogDebugWithLevel(fmt.Sprintf("Start run analysis %v/%v", key, len(listDocuments)), logger.DebugLevel)
		documentFindings = append(documentFindings, Run(units, rules)...)
	}
	return documentFindings
}

func breakTextUnitsIntoLimitOfUnit(allUnits []Unit, maxUnitsPerAnalysis int) (units [][]Unit) {
	units = [][]Unit{}
	startIndex := 0
	endIndex := maxUnitsPerAnalysis
	for i := 0; i < getTotalTextUnitsToRunByAnalysis(allUnits, maxUnitsPerAnalysis); i++ {
		units = append(units, []Unit{})
		units = toBreakUnitsAddUnitAndUpdateStartEndIndex(allUnits, units, startIndex, endIndex, i)
		startIndex = endIndex + 1
		endIndex += maxUnitsPerAnalysis
	}
	return units
}

func getTotalTextUnitsToRunByAnalysis(textUnits []Unit, maxUnitsPerAnalysis int) int {
	totalTextUnits := len(textUnits)
	if totalTextUnits <= maxUnitsPerAnalysis {
		return 1
	}
	totalUnitsToRun := float64(totalTextUnits / maxUnitsPerAnalysis)
	// nolint:staticcheck // is necessary usage pointless in math.ceil
	return int(math.Ceil(totalUnitsToRun))
}

func toBreakUnitsAddUnitAndUpdateStartEndIndex(
	allUnits []Unit, unitsToAppend [][]Unit, startIndex, endIndex, i int) [][]Unit {
	if len(allUnits[startIndex:]) <= endIndex {
		for k := range allUnits[startIndex:] {
			unitsToAppend[i] = append(unitsToAppend[i], allUnits[k])
		}
	} else {
		for k := range allUnits[startIndex:endIndex] {
			unitsToAppend[i] = append(unitsToAppend[i], allUnits[k])
		}
	}
	return unitsToAppend
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

//nolint:gomnd // improving in the feature
func createOutputFile(jsonFilePath string) (*os.File, error) {
	if _, err := os.Create(jsonFilePath); err != nil {
		return nil, err
	}
	outputFile, err := os.OpenFile(jsonFilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return outputFile, outputFile.Truncate(0)
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
