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

// +build !windows

package text

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestReadTextFileControlWithUTF8(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("examples", "csharp", "example2", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF8File, err := ReadTextFileUnix(filepath.Join("examples", "csharp", "example2", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF8File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF8File[:4])
	}
}

func TestReadTextFileControlWithUTF8WithBOM(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("examples", "csharp", "example2", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF8File, err := ReadTextFileUnix(filepath.Join("examples", "csharp", "example2", "PetsController.utf8bom.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF8File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF8File[:4])
	}
}

func TestReadTextFileWithUTF16LEWithBOM(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("examples", "csharp", "example2", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF16File, err := ReadTextFileUnix(filepath.Join("examples", "csharp", "example2", "PetsController.utf16lebom.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF16File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF16File[:4])
	}
}

func TestReadTextFileWithUTF16BEWithBOM(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile(filepath.Join("examples", "csharp", "example2", "PetsController.utf8.cs"))
	if err != nil {
		t.Fatal(err)
	}

	bytesReadFromUTF16File, err := ReadTextFileUnix(filepath.Join("examples", "csharp", "example2", "PetsController.utf16bebom.cs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(expectedBytes) != string(bytesReadFromUTF16File) {
		t.Errorf("Failed to read UTF16 LE: expected: %#v got %#v\n", string(expectedBytes)[:4], bytesReadFromUTF16File[:4])
	}
}
