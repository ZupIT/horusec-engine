package engine

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

type AdvisoryExample struct {
	ID string
	Name string
	Description string
}

func (a *AdvisoryExample) GetID() string { return a.ID }
func (a *AdvisoryExample) GetName() string { return a.Name }
func (a *AdvisoryExample) GetDescription() string { return a.Description }

func TestNewOutput(t *testing.T) {
	assert.IsType(t, NewOutput(nil), &Output{})
}

func TestOutput_ParseOutputReportToJSONFile(t *testing.T) {
	t.Run("Should not return error when getting value of findings", func(t *testing.T) {
		findings := []Finding{
			{
				ID:             uuid.New().String(),
				SourceLocation: Location{
					Filename: "/tmp.go",
					Line:     0,
					Column:   0,
				},
			},
		}
	 	res := NewOutput(findings).Value()
		assert.NotEmpty(t, res)
	})
	t.Run("Should not return empty when exists finding in advisory", func(t *testing.T) {
		ID := uuid.New().String()
		findings := []Finding{
			{
				ID:             ID,
				SourceLocation: Location{
					Filename: "/tmp.go",
					Line:     0,
					Column:   0,
				},
			},
		}
		advisoryExamples := []Advisory{
			&AdvisoryExample{
				ID:          ID,
				Name:        uuid.New().String(),
				Description: uuid.New().String(),
			},
		}
		assert.NotEmpty(t, NewOutput(findings).BuildReport(advisoryExamples))
	})
	t.Run("Should not return empty when not found finding in advisory", func(t *testing.T) {
		findings := []Finding{
			{
				ID:             uuid.New().String(),
				SourceLocation: Location{
					Filename: "/tmp.go",
					Line:     0,
					Column:   0,
				},
			},
		}
		advisoryExamples := []Advisory{
			&AdvisoryExample{
				ID:          uuid.New().String(),
				Name:        uuid.New().String(),
				Description: uuid.New().String(),
			},
		}
		assert.Empty(t, NewOutput(findings).BuildReport(advisoryExamples))
	})
	t.Run("Should not return error when exists finding in advisory and generate file with content", func(t *testing.T) {
		ID := uuid.New().String()
		outputFilePath := uuid.New().String() + "-tmp.json"
		findings := []Finding{
			{
				ID:             ID,
				SourceLocation: Location{
					Filename: "/tmp.go",
					Line:     0,
					Column:   0,
				},
			},
		}
		advisoryExamples := []Advisory{
			&AdvisoryExample{
				ID:          ID,
				Name:        uuid.New().String(),
				Description: uuid.New().String(),
			},
		}
		assert.NoError(t, NewOutput(findings).GenerateReportInOutputFilePath(advisoryExamples, outputFilePath))
		content, err := ioutil.ReadFile(outputFilePath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), ID)
		assert.NoError(t, os.RemoveAll(outputFilePath))
	})
	t.Run("Should not return error when exists finding in advisory and generate file with content empty", func(t *testing.T) {
		outputFilePath := uuid.New().String() + "-tmp.json"
		findings := []Finding{
			{
				ID:             uuid.New().String(),
				SourceLocation: Location{
					Filename: "/tmp.go",
					Line:     0,
					Column:   0,
				},
			},
		}
		advisoryExamples := []Advisory{
			&AdvisoryExample{
				ID:          uuid.New().String(),
				Name:        uuid.New().String(),
				Description: uuid.New().String(),
			},
		}
		assert.NoError(t, NewOutput(findings).GenerateReportInOutputFilePath(advisoryExamples, outputFilePath))
		content, err := ioutil.ReadFile(outputFilePath)
		assert.NoError(t, err)
		assert.Equal(t, string(content), "[]")
		assert.NoError(t, os.RemoveAll(outputFilePath))
	})
	t.Run("Should return error because path is invalid", func(t *testing.T) {
		outputPath := "./////"
		findings := []Finding{}
		assert.Error(t, NewOutput(findings).GenerateReportInOutputFilePath(nil, outputPath))
	})
}