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
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// regexNewLine regex representing the new line hexadecimal, equivalent of \n.
var regexNewLine = regexp.MustCompile("\x0a")

// File represents a file to be analyzed
type File struct {
	// AbsolutePath holds the complete path to the file (e.g. /home/user/myProject/router/handler.js)
	AbsolutePath         string
	RelativePath         string  // RelativePath holds the raw path relative to the root folder of the project
	Content              []byte  // Content holds all the file content
	Name                 string  // Name holds only the single name of the file (e.g. handler.js)
	newlineIndexes       [][]int // newlineIndexes holds information about where is the beginning and ending of each line
	newlineEndingIndexes []int   // newlineEndingIndexes represents the *start* index of each '\n' rune in the file
}

// NewTextFile create a new text file with all necessary info filled
func NewTextFile(relativeFilePath string, content []byte) (*File, error) {
	file := &File{
		RelativePath:   relativeFilePath,
		Content:        content,
		Name:           filepath.Base(relativeFilePath),
		newlineIndexes: regexNewLine.FindAllIndex(content, -1),
	}

	if err := file.setAbsFilePath(); err != nil {
		return nil, err
	}

	file.setNewlineEndingIndexes()

	return file, nil
}

// setAbsFilePath verifies if the filepath is absolute and set, otherwise it will parse and then set
func (f *File) setAbsFilePath() error {
	if filepath.IsAbs(f.RelativePath) {
		f.AbsolutePath = f.RelativePath

		return nil
	}

	absolutePath, err := filepath.Abs(f.RelativePath)
	f.AbsolutePath = absolutePath

	return err
}

// setNewlineEndingIndexes for each new line index set the ending index
func (f *File) setNewlineEndingIndexes() {
	for _, newlineIndex := range f.newlineIndexes {
		f.newlineEndingIndexes = append(f.newlineEndingIndexes, newlineIndex[0])
	}
}

// nolint:funlen,wsl // todo complex function need to be improved
// FindLineAndColumn get line and column using the beginning index of the example code
func (f *File) FindLineAndColumn(findingIndex int) (line, column int) {
	// findingIndex is the index of the beginning of the text we want to
	// locate inside the file

	// newlineEndingIndexes holds the indexes of each \n in the file
	// which means that each index in this slice is actually a line in the file

	// so we search for where we would put the findingIndex in the array
	// using a binary search algorithm, because this will give us the exact
	// lines that the index is between.
	lineIndex := f.binarySearch(findingIndex, f.newlineEndingIndexes)

	// Now with the right index found we have to get the previous \n
	// from the findingIndex, so it gets the right line
	if lineIndex < len(f.newlineEndingIndexes) {
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
		endOfCurrentLineInTheFile := f.newlineEndingIndexes[endOfCurrentLine]

		if lineIndex == 0 {
			column = findingIndex
		} else {
			column = (findingIndex - 1) - endOfCurrentLineInTheFile
		}
	}

	return line, column
}

// binarySearch function uses this search algorithm to find the index of the matching element.
func (f *File) binarySearch(searchIndex int, collection []int) (foundIndex int) {
	foundIndex = sort.Search(
		len(collection),
		func(index int) bool { return collection[index] >= searchIndex },
	)

	return
}

// nolint:funlen // todo complex function, needs to be improved
// ExtractSample search for the vulnerable code using the finding indexes
func (f *File) ExtractSample(findingIndex int) string {
	lineIndex := f.binarySearch(findingIndex, f.newlineEndingIndexes)

	if lineIndex < len(f.newlineEndingIndexes) {
		endOfPreviousLine := 0

		if lineIndex > 0 {
			endOfPreviousLine = f.newlineEndingIndexes[lineIndex-1] + 1
		}

		endOfCurrentLine := f.newlineEndingIndexes[lineIndex]
		lineContent := f.Content[endOfPreviousLine:endOfCurrentLine]

		return strings.TrimSpace(string(lineContent))
	}

	return ""
}
