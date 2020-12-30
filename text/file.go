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
	"github.com/ZupIT/horusec/development-kit/pkg/utils/logger"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

var (
	newlineFinder *regexp.Regexp = regexp.MustCompile("\x0a")

	PEMagicBytes   []byte = []byte{'\x4D', '\x5A'}                 // MZ
	ELFMagicNumber []byte = []byte{'\x7F', '\x45', '\x4C', '\x46'} // .ELF
)

const AcceptAllExtensions string = "**"

// binarySearch function uses this search algorithm to find the index of the matching element.
func binarySearch(searchIndex int, collection []int) (foundIndex int) {
	foundIndex = sort.Search(
		len(collection),
		func(index int) bool { return collection[index] >= searchIndex },
	)
	return
}

// TextFile represents a file to be analyzed
// nolint name is necessary for now called TextFile for not occurs breaking changes
type TextFile struct {
	DisplayName string // Holds the raw path relative to the root folder of the project
	Name        string // Holds only the single name of the file (e.g. handler.js)
	RawString   string // Holds all the file content

	// Holds the complete path to the file, could be absolute or not (e.g. /home/user/myProject/router/handler.js)
	PhysicalPath string

	// Indexes for internal file reference
	// newlineIndexes holds information about where is the beginning and ending of each line in the file
	newlineIndexes [][]int
	// newlineEndingIndexes represents the *start* index of each '\n' rune in the file
	newlineEndingIndexes []int
}

func NewTextFile(relativeFilePath string, content []byte) (TextFile, error) {
	formattedPhysicalPath, err := validateRelativeFilePath(relativeFilePath)
	if err != nil {
		return TextFile{}, err
	}

	return createTextFileByPath(formattedPhysicalPath, relativeFilePath, content), nil
}

func createTextFileByPath(formattedPhysicalPath, relativeFilePath string, content []byte) TextFile {
	_, formattedFilename := filepath.Split(formattedPhysicalPath)
	textfile := TextFile{
		PhysicalPath: formattedPhysicalPath,
		RawString:    string(content),

		// Display info
		Name:        formattedFilename,
		DisplayName: relativeFilePath,
	}
	textfile.newlineIndexes = newlineFinder.FindAllIndex(content, -1)

	for _, newlineIndex := range textfile.newlineIndexes {
		textfile.newlineEndingIndexes = append(textfile.newlineEndingIndexes, newlineIndex[0])
	}
	return textfile
}

func validateRelativeFilePath(relativeFilePath string) (string, error) {
	if !filepath.IsAbs(relativeFilePath) {
		return filepath.Abs(relativeFilePath)
	}

	return relativeFilePath, nil
}

func (textfile TextFile) Content() string {
	return textfile.RawString
}

// nolint TODO: Remove commentaries and refactor method to clean code
func (textfile TextFile) FindLineAndColumn(findingIndex int) (line, column int) {
	// findingIndex is the index of the beginning of the text we want to
	// locate inside the file

	// newlineEndingIndexes holds the indexes of each \n in the file
	// which means that each index in this slice is actually a line in the file

	// so we search for where we would put the findingIndex in the array
	// using a binary search algorithm, because this will give us the exact
	// lines that the index is between.
	lineIndex := binarySearch(findingIndex, textfile.newlineEndingIndexes)

	// Now with the right index found we have to get the previous \n
	// from the findingIndex, so it gets the right line
	if lineIndex < len(textfile.newlineEndingIndexes) {
		// we add +1 here because we want the line to
		// reflect the "human" line count, not the indexed one in the slice
		line = lineIndex + 1

		endOfCurrentLine := lineIndex - 1

		// If there is no previous line the finding is in the beginning
		// of the file, so we just normalize to avoid signing issues (like -1 messing the indexing)
		if endOfCurrentLine <= 0 {
			endOfCurrentLine = 0
		}

		// now we access the textual index in the slice to ge the column
		endOfCurrentLineInTheFile := textfile.newlineEndingIndexes[endOfCurrentLine]
		if findingIndex == 0 {
			column = endOfCurrentLineInTheFile
		} else {
			column = (findingIndex - 1) - endOfCurrentLineInTheFile
		}
	}
	return line, column
}

func (textfile TextFile) ExtractSample(findingIndex int) string {
	lineIndex := binarySearch(findingIndex, textfile.newlineEndingIndexes)

	if lineIndex < len(textfile.newlineEndingIndexes) {
		endOfPreviousLine := 0
		if lineIndex > 0 {
			endOfPreviousLine = textfile.newlineEndingIndexes[lineIndex-1] + 1
		}
		endOfCurrentLine := textfile.newlineEndingIndexes[lineIndex]

		lineContent := textfile.RawString[endOfPreviousLine:endOfCurrentLine]

		return strings.TrimSpace(lineContent)
	}

	return ""
}

func ReadAndCreateTextFile(filename string) (TextFile, error) {
	var textFileContent []byte
	var err error
	if runtime.GOOS == "windows" {
		textFileContent, err = ReadTextFileWin(filename)
	} else {
		textFileContent, err = ReadTextFileUnix(filename)
	}
	if err != nil {
		return TextFile{}, err
	}

	textFileMagicBytes := textFileContent[:4]
	if bytes.Equal(textFileMagicBytes, ELFMagicNumber) {
		// Ignore Linux binaries
		return TextFile{}, nil
	} else if bytes.Equal(textFileContent[:2], PEMagicBytes) {
		// Ignore Windows binaries
		return TextFile{}, nil
	}

	return NewTextFile(filename, textFileContent)
}

// The Param extensionAccept is an filter to check if you need get textUnit for file with this extesion
//   Example: []string{".java"}
// If an item of slice contains is equal the "**" it's will accept all extensions
//   Example: []string{"**"}
func LoadDirIntoSingleUnit(path string, extensionsAccept []string) (TextUnit, error) {
	listTextUnit, err := loadDirIntoUnit(path, 0, extensionsAccept)
	if err != nil {
		return TextUnit{}, err
	}
	if len(listTextUnit) < 1 {
		return TextUnit{}, nil
	}
	return listTextUnit[0], nil
}

// The Param extensionAccept is an filter to check if you need get textUnit for file with this extesion
//   Example: []string{".java"}
// If an item of slice contains is equal the "**" it's will accept all extensions
//   Example: []string{"**"}
// nolint Complex method for pass refactor now TODO: Refactor this method in the future to clean code
func LoadDirIntoMultiUnit(path string, maxFilesPerTextUnit int, extensionsAccept []string) ([]TextUnit, error) {
	return loadDirIntoUnit(path, maxFilesPerTextUnit, extensionsAccept)
}

func loadDirIntoUnit(path string, maxFilesPerTextUnit int, extensionsAccept []string) ([]TextUnit, error) {
	filesToRun, err := getFilesPathIntoProjectPath(path, extensionsAccept)
	if err != nil {
		return []TextUnit{}, err
	}
	return getTextUnitsFromFilesPath(filesToRun, maxFilesPerTextUnit)
}

func getFilesPathIntoProjectPath(projectPath string, extensionsAccept []string) (filesToRun []string, err error) {
	return filesToRun, filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if checkIfEnableExtension(path, extensionsAccept) {
				filesToRun = append(filesToRun, path)
			}
		}
		return nil
	})
}

func getTextUnitsFromFilesPath(filesToRun []string, maxFilesPerTextUnit int) (textUnits []TextUnit, err error) {
	textUnits = []TextUnit{{}}
	lastIndexToAdd := 0
	for k, currentFile := range filesToRun {
		time.Sleep(15 * time.Millisecond)
		currentTime := time.Now()
		textUnits, lastIndexToAdd, err = readFileAndExtractTextUnit(
			textUnits, lastIndexToAdd, maxFilesPerTextUnit, currentFile)
		logger.LogTraceWithLevel(fmt.Sprintf(
			"Read file in %v Microseconds. Total files read: %v/%v ",
			time.Since(currentTime).Microseconds(), k, len(filesToRun)), logger.TraceLevel, currentFile)
		if err != nil {
			return []TextUnit{}, err
		}
	}
	return textUnits, nil
}

func readFileAndExtractTextUnit(
	textUnits []TextUnit, lastIndexToAdd, maxFilesPerTextUnit int, currentFile string) ([]TextUnit, int, error) {
	textFile, err := ReadAndCreateTextFile(currentFile)
	if err != nil {
		return []TextUnit{}, lastIndexToAdd, err
	}
	textUnits[lastIndexToAdd].Files = append(textUnits[lastIndexToAdd].Files, textFile)
	if maxFilesPerTextUnit > 0 && len(textUnits[lastIndexToAdd].Files) >= maxFilesPerTextUnit {
		textUnits = append(textUnits, TextUnit{})
		return textUnits, lastIndexToAdd + 1, nil
	}
	return textUnits, lastIndexToAdd, nil
}

func checkIfEnableExtension(path string, extensionsAccept []string) bool {
	ext := filepath.Ext(path)
	for _, extAccept := range extensionsAccept {
		if ext == extAccept || extAccept == AcceptAllExtensions {
			return true
		}
	}
	return false
}
