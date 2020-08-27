package engine

import (
	"encoding/json"
	"os"
)

type Interface interface {
	ParseOutputReportToJSONFile(report []Report, outputFilePath string) error
}

type Output struct{}

func NewOutput() Interface {
	return &Output{}
}

func (o *Output) ParseOutputReportToJSONFile(report []Report, outputFilePath string) error {
	if report == nil {
		report = []Report{}
	}
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
