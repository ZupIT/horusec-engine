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
	"errors"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	sampleKotlin = `
		package org.jetbrains.kotlin.demo

		import org.springframework.web.bind.annotation.GetMapping
		import org.springframework.web.bind.annotation.RequestParam
		import org.springframework.web.bind.annotation.RestController
		import java.util.concurrent.atomic.AtomicLong
		
		@RestController
		class GreetingController {
		
			val counter = AtomicLong()
		
			@GetMapping("/greeting")
			fun greeting(@RequestParam(value = "name", defaultValue = "World") name: String) =
					Greeting(counter.incrementAndGet(), "Hello, $name")
		
		}
	`

	sampleGo = `
		package version
		
		import (
			"github.com/ZupIT/horusec/development-kit/pkg/utils/logger"
			"github.com/spf13/cobra"
		)
		
		type IVersion interface {
			CreateCobraCmd() *cobra.Command
		}
		
		type Version struct {
		}
		
		func NewVersionCommand() IVersion {
			return &Version{}
		}
		
		func (v *Version) CreateCobraCmd() *cobra.Command {
			return &cobra.Command{
				Use:     "version",
				Short:   "Actual version installed of the horusec",
				Example: "horusec version",
				RunE: func(cmd *cobra.Command, args []string) error {
					logger.LogPrint(cmd.Short + " is: ")
					return nil
				},
			}
		}
	`

	sampleJs = `
		const http = require('http');
		
		const hostname = '127.0.0.1';
		const port = 3000;
		
		const server = http.createServer((req, res) => {
		  res.statusCode = 200;
		  res.setHeader('Content-Type', 'text/plain');
		  res.end('Hello World');
		});
		
		server.listen(port, hostname, () => {
		  console.log('Server running at https://${hostname}:${port}/'');
		});
`
)

func getFindingIndex(sample, expression string) (int, error) {
	indexes := regexp.MustCompile(expression).FindIndex([]byte(sample))
	if len(indexes) > 0 {
		return indexes[0], nil
	}

	return 0, errors.New("failed to get finding indexes")
}

func TestFindLineAndColumn(t *testing.T) {
	testCases := []struct {
		name            string
		regexExpression string
		codeSample      string
		expectedLine    int
		expectedColumn  int
	}{
		{
			name:            "Should success find line and column for kotlin",
			regexExpression: `name\:`,
			codeSample:      sampleKotlin,
			expectedLine:    15,
			expectedColumn:  70,
		},
		{
			name:            "Should success find line and column for go",
			regexExpression: `cmd\.Short`,
			codeSample:      sampleGo,
			expectedLine:    26,
			expectedColumn:  21,
		},
		{
			name:            "Should success find line and column for js",
			regexExpression: `server\.listen`,
			codeSample:      sampleJs,
			expectedLine:    13,
			expectedColumn:  2,
		},
	}

	for index, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			findingIndex, err := getFindingIndex(testCase.codeSample, testCase.regexExpression)
			assert.NoError(t, err)

			file, err := NewTextFile("test", []byte(testCase.codeSample))
			assert.NoError(t, err)

			line, column := file.FindLineAndColumn(findingIndex)
			assert.Equalf(t, testCase.expectedLine, line, "failed to find correct line, test case: %d", index)
			assert.Equalf(t, testCase.expectedColumn, column, "failed to find correct column, test case: %d", index)
		})
	}
}

func TestExtractSample(t *testing.T) {
	testCases := []struct {
		name               string
		regexExpression    string
		codeSample         string
		expectedCodeSample string
	}{
		{
			name:               "Should success find code sample for kotlin",
			regexExpression:    `name\:`,
			codeSample:         sampleKotlin,
			expectedCodeSample: "fun greeting(@RequestParam(value = \"name\", defaultValue = \"World\") name: String) =",
		},
		{
			name:               "Should success find code sample for go",
			regexExpression:    `cmd\.Short`,
			codeSample:         sampleGo,
			expectedCodeSample: "logger.LogPrint(cmd.Short + \" is: \")",
		},
		{
			name:               "Should success find code sample for js",
			regexExpression:    `server\.listen`,
			codeSample:         sampleJs,
			expectedCodeSample: "server.listen(port, hostname, () => {",
		},
	}

	for index, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			findingIndex, err := getFindingIndex(testCase.codeSample, testCase.regexExpression)
			assert.NoError(t, err)

			file, err := NewTextFile("test", []byte(testCase.codeSample))
			assert.NoError(t, err)

			sample := file.ExtractSample(findingIndex)
			assert.Equalf(t, testCase.expectedCodeSample, sample, "failed to find code sample, test case: %d", index)
		})
	}
}

func TestNewTextFile(t *testing.T) {
	t.Run("Should success create a new text file", func(t *testing.T) {
		relativeFilePath := filepath.Join("examples", "go", "example1", "api", "server.go")

		file, err := NewTextFile(relativeFilePath, []byte(sampleGo))
		assert.NoError(t, err)

		assert.Truef(t, filepath.IsAbs(file.AbsolutePath), "path it's not absolute")
		assert.Equalf(t, relativeFilePath, file.RelativePath, "failed to match relative path")
		assert.Equalf(t, sampleGo, string(file.Content), "failed to match content")
		assert.Equalf(t, filepath.Base(relativeFilePath), file.Name, "failed to match file name")
		assert.Lenf(t, file.newlineIndexes, 30, "sample go contains 30 new line indexes")
		assert.Lenf(t, file.newlineEndingIndexes, 30, "sample go contains 30 ending indexes")
	})
}
