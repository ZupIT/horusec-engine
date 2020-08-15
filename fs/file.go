package fs

import (
	"path/filepath"
	"regexp"
	"sort"
)

var (
	newlineFinder *regexp.Regexp = regexp.MustCompile("\x0a")
)

// binarySearch function uses this search algorithm to find the index of the matching element.
func binarySearch(searchIndex int, collection []int) (foundIndex int) {
	foundIndex = sort.Search(
		len(collection),
		func(index int) bool { return collection[index] >= searchIndex },
	)
	return
}

type File interface {
	Content() string
	FindLineAndColumn(findingIndex int) (line int, column int)
}

// TextFile represents a file to be analyzed
type TextFile struct {
	DisplayName string // Holds only the single name of the file (e.g. handler.js)
	RawString   string // Holds all the file content

	// Holds the complete path to the file, could be absolute or not (e.g. /home/user/Documents/myProject/router/handler.js)
	PhysicalPath string

	// Indexes for internal file reference
	// newlineIndexes holds information about where is the beginning and ending of each line in the file
	newlineIndexes [][]int
	// newlineEndingIndexes represents the *start* index of each '\n' rune in the file
	newlineEndingIndexes []int
}

func NewTextFile(filename string, content []byte) (*TextFile, error) {
	var err error
	var formattedPhysicalPath string

	if !filepath.IsAbs(filename) {
		formattedPhysicalPath, err = filepath.Abs(filename)

		if err != nil {
			return nil, err
		}
	}

	_, formattedFilename := filepath.Split(formattedPhysicalPath)

	textfile := &TextFile{
		PhysicalPath: formattedPhysicalPath,
		RawString:    string(content),

		// Display info
		DisplayName: formattedFilename,
	}

	textfile.newlineIndexes = newlineFinder.FindAllIndex(content, -1)

	for _, newlineIndex := range textfile.newlineIndexes {
		textfile.newlineEndingIndexes = append(textfile.newlineEndingIndexes, newlineIndex[0])
	}

	return textfile, nil
}

func (textfile *TextFile) Content() string {
	return textfile.RawString
}

func (textfile *TextFile) FindLineAndColumn(findingIndex int) (line, column int) {
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
		column = (findingIndex - 1) - endOfCurrentLineInTheFile
	}

	return
}
