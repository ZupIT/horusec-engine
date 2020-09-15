package text

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	newlineFinder *regexp.Regexp = regexp.MustCompile("\x0a")
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
type TextFile struct {
	DisplayName string // Holds the raw path relative to the root folder of the project
	Name        string // Holds only the single name of the file (e.g. handler.js)
	RawString   string // Holds all the file content

	// Holds the complete path to the file, could be absolute or not (e.g. /home/user/Documents/myProject/router/handler.js)
	PhysicalPath string

	// Indexes for internal file reference
	// newlineIndexes holds information about where is the beginning and ending of each line in the file
	newlineIndexes [][]int
	// newlineEndingIndexes represents the *start* index of each '\n' rune in the file
	newlineEndingIndexes []int
}

func NewTextFile(relativeFilePath string, content []byte) (TextFile, error) {
	var err error
	var formattedPhysicalPath string

	if !filepath.IsAbs(relativeFilePath) {
		formattedPhysicalPath, err = filepath.Abs(relativeFilePath)

		if err != nil {
			return TextFile{}, err
		}
	} else {
		formattedPhysicalPath = relativeFilePath
	}

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

	return textfile, nil
}

func (textfile TextFile) Content() string {
	return textfile.RawString
}

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
		column = (findingIndex - 1) - endOfCurrentLineInTheFile
	}

	return
}

func (textfile TextFile) ExtractSample(findingIndex int) string {
	lineIndex := binarySearch(findingIndex, textfile.newlineEndingIndexes)
	endOfPreviousLineIndex := 0
	if lineIndex > 0 {
		endOfPreviousLineIndex = lineIndex - 1
	}

	if lineIndex < len(textfile.newlineEndingIndexes) {
		endOfPreviousLine := textfile.newlineEndingIndexes[endOfPreviousLineIndex]
		endOfCurrentLine := textfile.newlineEndingIndexes[lineIndex]

		lineContent := textfile.RawString[endOfPreviousLine+1 : endOfCurrentLine]

		return strings.TrimSpace(lineContent)
	}

	return ""
}

func ReadAndCreateTextFile(filename string) (TextFile, error) {
	textFileContent, err := ReadTextFile(filename)

	if err != nil {
		return TextFile{}, err
	}

	return NewTextFile(filename, textFileContent)
}

// The Param extensionAccept is an filter to check if you need get textUnit for file with this extesion
//   Example: []string{".java"}
// If an item of slice contains is equal the "**" it's will accept all extensions
//   Example: []string{"**"}
func LoadDirIntoSingleUnit(path string, extensionsAccept []string) (TextUnit, error) {
	unit := TextUnit{}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		textFile, err := validateAndGetTextFileByPath(path, extensionsAccept)
		if err != nil || textFile == nil {
			return err
		}
		unit.Files = append(unit.Files, *textFile)
		return nil
	})
	return unit, err
}

func validateAndGetTextFileByPath(path string, extensionsAccept []string) (*TextFile, error) {
	if checkIfEnableExtension(path, extensionsAccept) {
		file, err := getTextFileByPath(path)
		if err != nil {
			return nil, err
		}
		return &file, nil
	}
	return nil, nil
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

func getTextFileByPath(path string) (TextFile, error) {
	file, err := ReadAndCreateTextFile(path)
	if err != nil {
		return TextFile{}, err
	}
	return file, nil
}
