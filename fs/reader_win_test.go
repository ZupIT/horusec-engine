// +build windows

package fs

import (
	"path/filepath"
	"testing"
)

func TestWinReadTextFileShouldFailWithUTF16LEWithoutBOM(t *testing.T) {
	_, err := ReadTextFile(filepath.Join("samples", "PetsController.utf16le.cs"))

	if err == nil {
		t.Fatalf("Should have returned error for files encoded with UTF16 LE without BOM")
	}
}

func TestWinReadTextFileShouldFailWithUTF16BEWithoutBOM(t *testing.T) {
	_, err := ReadTextFile(filepath.Join("samples", "PetsController.utf16be.cs"))

	if err == nil {
		t.Fatalf("Should have returned error for files encoded with UTF16 BE without BOM")
	}
}
