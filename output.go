package engine

import (
	"encoding/json"
	"os"
)

type IOutput interface {
	Value() []Finding
	BuildReport(advisories []Advisory) []Report
	GenerateReportInOutputFilePath(advisories []Advisory, outputFilePath string) error
}

type Report struct {
	ID             string   // Comes from Advisory::GetID/0
	Name           string   // Comes from Advisory::GetName/0
	Description    string   // Comes from Advisory::GetDescription/0
	SourceLocation Location // Comes from the Finding
}

type Output struct{
	findings []Finding
}

func NewOutput(findings []Finding) IOutput {
	return &Output{
		findings: findings,
	}
}

func (o *Output) Value() []Finding {
	return o.findings
}

func (o *Output) BuildReport(advisories []Advisory) (programReport []Report) {
	for _, advisory := range advisories {
		for _, finding := range o.findings {
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

func (o *Output) GenerateReportInOutputFilePath(advisories []Advisory, outputFilePath string) error {
	report := o.BuildReport(advisories)
	bytesToWrite, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return o.parseFilePathToAbsAndCreateOutputJSON(bytesToWrite, outputFilePath)
}

func (o *Output) parseFilePathToAbsAndCreateOutputJSON(bytesToWrite []byte, outputFilePath string) error {
	if _, err := os.Create(outputFilePath); err != nil {
		return err
	}
	return o.openJSONFileAndWriteBytes(bytesToWrite, outputFilePath)
}

func (o *Output) openJSONFileAndWriteBytes(bytesToWrite []byte, completePath string) error {
	outputFile, err := os.OpenFile(completePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	if err = outputFile.Truncate(0); err != nil {
		return err
	}
	if bytesWritten, err := outputFile.Write(bytesToWrite); err != nil || bytesWritten != len(bytesToWrite) {
		return err
	}
	return nil
}
