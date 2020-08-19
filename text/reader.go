// +build !windows

package text

import (
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// newUnicodeReader creates a transformer to read UTF16 LE or BE MS-Windows files
// essentially transforming everything to UTF-8, if and only if the file have the BOM
func newUnicodeReader(defaultReader io.Reader) io.Reader {
	decoder := unicode.UTF8.NewDecoder()
	return transform.NewReader(defaultReader, unicode.BOMOverride(decoder))
}

// ReadTextFile reads the content of a file, converting when possible
// the encoding to UTF-8.
func ReadTextFile(filename string) ([]byte, error) {
	fileDescriptor, err := os.Open(filename)

	if err != nil {
		return []byte{}, err
	}

	defer fileDescriptor.Close()

	reader := newUnicodeReader(fileDescriptor)

	utf8FormattedString, err := ioutil.ReadAll(reader)

	if err != nil {
		return []byte{}, err
	}

	return utf8FormattedString, nil
}
