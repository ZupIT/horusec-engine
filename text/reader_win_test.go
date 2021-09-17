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

// +build windows

package text

import (
	"path/filepath"
	"testing"
)

func TestWinReadTextFileShouldFailWithUTF16LEWithoutBOM(t *testing.T) {
	_, err := ReadTextFileWin(filepath.Join(samples, "PetsController.utf16le.cs"))

	if err == nil {
		t.Fatalf("Should have returned error for files encoded with UTF16 LE without BOM")
	}
}

func TestWinReadTextFileShouldFailWithUTF16BEWithoutBOM(t *testing.T) {
	_, err := ReadTextFileWin(filepath.Join(samples, "PetsController.utf16be.cs"))

	if err == nil {
		t.Fatalf("Should have returned error for files encoded with UTF16 BE without BOM")
	}
}
