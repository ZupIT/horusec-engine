// +build !windows

package text

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestReadTextFileControlWithUTF8(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("samples", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF8File, err := ReadTextFile(filepath.Join("samples", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF8File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF8File[:4])
	}
}

func TestReadTextFileControlWithUTF8WithBOM(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("samples", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF8File, err := ReadTextFile(filepath.Join("samples", "PetsController.utf8bom.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF8File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF8File[:4])
	}
}

func TestReadTextFileWithUTF16LEWithBOM(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("samples", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF16File, err := ReadTextFile(filepath.Join("samples", "PetsController.utf16lebom.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF16File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF16File[:4])
	}
}

func TestReadTextFileWithUTF16BEWithBOM(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("samples", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF16File, err := ReadTextFile(filepath.Join("samples", "PetsController.utf16bebom.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF16File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF16File[:4])
	}
}
