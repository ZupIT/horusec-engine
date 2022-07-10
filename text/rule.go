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

package text

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/bmatcuk/doublestar"

	engine "github.com/ZupIT/horusec-engine"
)

// MatchType represents the possibles match types of the engine
type MatchType int

const (
	// OrMatch for each regex that match will report a vulnerability
	OrMatch MatchType = iota

	// Regular do the exact same thing as OrMatch, will be depreciated in the future to simplify engine use
	Regular

	// NotMatch will report any file that don't match the regex expressions
	NotMatch

	// AndMatch need that all regex expressions match to report the vulnerability, it will get the first regex expression
	// the use as base to the reported vulnerability
	AndMatch
)

// peMagicBytes hexadecimal used to find windows binaries
// elfMagicNumber hexadecimal used to find linux binaries
var (
	peMagicBytes   = []byte{'\x4D', '\x5A'}                 // MZ
	elfMagicNumber = []byte{'\x7F', '\x45', '\x4C', '\x46'} // .ELF
)

// Rule represents the vulnerability that should be searched in the file. It contains some predefined information about
// the vulnerability like the id, name, description, severity, confidence, match type that should be applied and the
// regular expressions used to match the vulnerable code
type Rule struct {
	engine.Metadata
	Type        MatchType
	Expressions []*regexp.Regexp
}

// Run start a static code analysis using regular expressions, it will read the file content as bytes and create a text
// file with it. The text file contains all information needed to find the vulnerable code when the regular expressions
// match. There's also a validation to ignore binary files
func (r *Rule) Run(path string) ([]engine.Finding, error) {
	content, err := r.getFilteredFileContent(path)
	if content == nil || err != nil {
		return nil, nil
	}

	if r.isBinary(content) {
		return nil, nil
	}

	textFile, err := NewTextFile(path, content)
	if err != nil {
		return nil, err
	}

	return r.runByRuleType(textFile)
}

func (r *Rule) getFilteredFileContent(path string) ([]byte, error) {
	matched, _ := doublestar.Match(r.Filter, path)
	if !matched {
		return nil, nil
	}

	content, err := r.getFileContent(path)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// getFileContent opens the file using the file path, reads and returns its contents as bytes. After all done closes
// the file
func (r *Rule) getFileContent(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// runByRuleType determines which match type should be applied and ran according the rule
func (r *Rule) runByRuleType(file *File) ([]engine.Finding, error) {
	switch r.Type {
	// TODO: regular type do the exact same thing as OrMatch, will be depreciated in the future to simplify engine use
	case OrMatch, Regular:
		return r.runOrMatch(file)
	case NotMatch:
		return r.runNotMatch(file)
	case AndMatch:
		return r.runAndMatch(file)
	}

	return nil, fmt.Errorf("invalid rule type")
}

// runNotMatch for each regex expression will search for matches in the file and return they index and create the
// findings with them. Different of the other types, this type will report files that didn't have any match with each
// one of the regex expressions.
// TODO: since this match type search for files that didn't match the rules, we can't get a sample code,
// line and column, witch lead to a really vague report. Need to be revisited and improved in the future.
func (r *Rule) runNotMatch(file *File) ([]engine.Finding, error) {
	var findings []engine.Finding

	for _, expression := range r.Expressions {
		if expression.FindAllIndex(file.Content, -1) == nil {
			findings = append(findings, r.newFinding(file.RelativePath, "", 0, 0))
		}
	}

	return findings, nil
}

// runAndMatch for each regex expression will search for matches in the file and return they index and create the
// findings with them. If any of the regex expressions don't match, it should return nil, all regex expressions should
// match to be a valid vulnerability. In case of all have matched the first finding will be returned to be used to
// generate the report
//nolint:funlen // necessary length, it's not a complex func, maybe can be improved in the future
func (r *Rule) runAndMatch(file *File) ([]engine.Finding, error) {
	var findings []engine.Finding

	isFailedToMatchAll := false

	for _, expression := range r.Expressions {
		findingIndexes := expression.FindAllIndex(file.Content, -1)
		if findingIndexes != nil {
			findings = append(findings, r.createFindingsFromIndexes(findingIndexes, file)...)

			continue
		}

		isFailedToMatchAll = true

		break
	}

	return r.getFirstFindingIfAllMatched(isFailedToMatchAll, findings), nil
}

// getFirstFindingIfAllMatched checks if all regex expressions matched, if not will return nil. In case of all of them
// have match will get the first finding of the slice and return it to be used to generate the report
func (r *Rule) getFirstFindingIfAllMatched(isFailedToMatchAll bool, findings []engine.Finding) []engine.Finding {
	if isFailedToMatchAll {
		return nil
	}

	if len(findings) >= 1 {
		return []engine.Finding{findings[0]}
	}

	return nil
}

// runOrMatch for each regex expression will search for matches in the file and return they index and create the
// findings with them. Since the OrMatch type rules can match many times, they can return more than one finding for rule
func (r *Rule) runOrMatch(file *File) ([]engine.Finding, error) {
	var findings []engine.Finding

	for _, expression := range r.Expressions {
		findingIndexes := expression.FindAllIndex(file.Content, -1)
		if findingIndexes != nil {
			findings = append(findings, r.createFindingsFromIndexes(findingIndexes, file)...)

			continue
		}
	}

	return findings, nil
}

// createFindingsFromIndexes for each index found of a possible vulnerability will get the line, column and code sample
// and create a new finding to append into the result
func (r *Rule) createFindingsFromIndexes(findingIndexes [][]int, file *File) (findings []engine.Finding) {
	for _, findingIndex := range findingIndexes {
		line, column := file.FindLineAndColumn(findingIndex[0])
		codeSample := file.ExtractSample(findingIndex[0])

		findings = append(findings, r.newFinding(
			file.RelativePath,
			codeSample,
			line,
			column,
		))
	}

	return findings
}

// newFinding create a new finding with the information of the vulnerability obtained from the file
func (r *Rule) newFinding(filename, codeSample string, line, column int) engine.Finding {
	return engine.Finding{
		ID:          r.ID,
		Name:        r.Name,
		Severity:    r.Severity,
		Confidence:  r.Confidence,
		Description: r.Description,
		CodeSample:  codeSample,
		SourceLocation: engine.Location{
			Filename: filename,
			Line:     line,
			Column:   column,
		},
	}
}

// isBinary verify if the file being analyzed is a binary file
func (r *Rule) isBinary(content []byte) bool {
	// Ignore Linux binaries
	if bytes.Equal(content[:4], elfMagicNumber) {
		return true
	}

	// Ignore Windows binaries
	if bytes.Equal(content[:2], peMagicBytes) {
		return true
	}

	return false
}
