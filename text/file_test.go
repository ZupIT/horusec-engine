package text

import (
	"regexp"
	"testing"
)

func TestFindLineAndColumnWithAKotlinController(t *testing.T) {
	var exampleKotlinController = `package org.jetbrains.kotlin.demo

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

	nameVariableLine := 14
	nameVariableColumn := 71

	// So we lookup for something in the file
	// in this case the 'name: String' variable
	// and the method should return the correct line and column
	// for where it is, in a human readable form.
	nameStringExtractor := regexp.MustCompile(`name\:`)

	findingIndex := nameStringExtractor.FindStringIndex(exampleKotlinController)

	controllerTextFile, err := NewTextFile("example/controller.kt", []byte(exampleKotlinController))

	if err != nil {
		t.Error(err)
	}

	line, column := controllerTextFile.FindLineAndColumn(findingIndex[0])

	if line != nameVariableLine || column != nameVariableColumn {
		t.Errorf(
			"Failed to find the right line and column. Wanted: %d:%d. Found: %d:%d",
			nameVariableLine, nameVariableColumn,
			line, column,
		)
	}
}

func TestFindLineAndColumnWithAGoFile(t *testing.T) {
	var exampleGoFile = `package version

import (
	"github.com/ZupIT/horus/development-kit/pkg/utils/logger"
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
		Short:   "Actual version installed of the horus",
		Example: "horus version",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.LogPrint(cmd.Short + " is: ")
			return nil
		},
	}
}
`
	cmdShortVariableLine := 25
	cmdShortVariableColumn := 19

	cmdShortExtractor := regexp.MustCompile(`cmd\.Short`)

	findingIndex := cmdShortExtractor.FindStringIndex(exampleGoFile)

	goTextFile, err := NewTextFile("example/cmd/version.go", []byte(exampleGoFile))

	if err != nil {
		t.Error(err)
	}

	line, column := goTextFile.FindLineAndColumn(findingIndex[0])

	if line != cmdShortVariableLine || column != cmdShortVariableColumn {
		t.Errorf(
			"Failed to find the right line and column. Wanted: %d:%d. Found: %d:%d",
			cmdShortVariableLine, cmdShortVariableColumn,
			line, column,
		)
	}
}

func TestNameFormattingAndDisplaying(t *testing.T) {
	expectedName := "version.go"

	goTextFile, err := NewTextFile("example/cmd/version.go", []byte{})

	if err != nil {
		t.Error(err)
	}

	if goTextFile.Name != expectedName {
		t.Errorf(
			"Failed to format the Name of the TextFile for Golang. Wanted: %s, Got: %s",
			expectedName,
			goTextFile.Name,
		)
	}
}
